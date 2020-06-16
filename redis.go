package main

import (
	"encoding/json"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/brianbroderick/logit"
	"github.com/garyburd/redigo/redis"
)

var (
	pool          *redis.Pool
	redisPassword string
)

func getLog(redisKey string) (*query, error) {
	var data json.RawMessage

	conn := pool.Get()
	defer conn.Close()

	reply, err := conn.Do("LPOP", redisKey)
	if err != nil {
		return nil, err
	}

	if reply != nil {
		data, err = redis.Bytes(reply, err)
		if err != nil {
			return nil, err
		}

		query, suppressed, err := newQuery(data, redisKey)
		if err != nil {
			return nil, err
		}
		if suppressed {
			return nil, nil
		}
		return query, nil
	}
	return nil, nil
}

func startRedisBatch(redisKey string, task string) {
	var (
		keyLog            string
		lastLog           int
		lastProcessed     int64
		msgLog            string
		nap               int
		processedMessages int64
		sleepDuration     time.Duration
	)

	for {
		ok, msgCount, queueLength, err := getMultiLog(redisKey, task)
		if err != nil {
			logit.Error("Error in getMultiLog: %e", err.Error())
		}
		if ok == false {
			sleepDuration = getDuration(nap)

			nap += int(sleepDuration)
			time.Sleep((time.Second * sleepDuration))

			// idle for 20 seconds, emit message
			if (nap < 60 && nap-lastLog >= 30) || (nap >= 60 && nap-lastLog >= 60) {
				logit.Info("%s %s %s", yellow(nap), white(" seconds since received data from "), yellow(redisKey))
				lastLog = nap
				processedMessages = 0
			}
		} else {
			nap = 0
			lastLog = 0
			processedMessages += msgCount
			if (processedMessages < 1000000 && processedMessages-lastProcessed >= 10000) ||
				(processedMessages >= 1000000 && processedMessages-lastProcessed >= 100000) {
				keyLog = colorKey(redisKey, processedMessages)
				msgLog = colorKey(strconv.FormatInt(processedMessages, 10), processedMessages)
				logit.Info("%s messages processed from %s since last reset", msgLog, keyLog)
				logit.Info("Current queue length for %s is %s", keyLog, red(queueLength))
				lastProcessed = processedMessages
			}
		}
	}
}

func colorKey(str string, counter int64) string {
	var colorString string
	if counter > 100000000 {
		colorString = blue(str)
	} else if counter > 20000000 {
		colorString = green(str)
	} else if counter > 2000000 {
		colorString = cyan(str)
	} else {
		colorString = magenta(str)
	}
	return colorString
}

// task == "query" for normal operations.
// task == "stats" to handle a stats payload. See stats.go for more details

func getMultiLog(redisKey string, task string) (bool, int64, int64, error) {
	conn := pool.Get()
	defer conn.Close()

	// get list length
	l, err := conn.Do("LLEN", redisKey)
	if err != nil {
		return true, 0, 0, err
	}

	llen, err := redis.Int64(l, err)
	if err != nil {
		return true, 0, 0, err
	}

	queueLength := llen

	// LPOP at most 100 at a time
	if llen > 100 {
		llen = 100
	}

	if llen > 0 {
		conn.Send("MULTI")
		var n int64
		for n = 0; n < llen; n++ {
			conn.Send("LPOP", redisKey)
		}

		reply, err := redis.Values(conn.Do("EXEC"))
		if err != nil {
			return true, 0, queueLength, err
		}

		for _, datum := range reply {
			if datum != nil {
				q, err := redis.Bytes(datum, err)
				if err != nil {
					return true, 0, queueLength, err
				}

				if task == "query" {
					query, suppressed, err := newQuery(q, redisKey)
					if err != nil {
						logit.Error("Error in newQuery: %e", err.Error())
					} else if suppressed {
						// no-op
					} else {
						addToQueries(roundToMinute(query.timestamp), query)
					}
				} else {
					stats, err := newStats(q, redisKey)
					if err != nil {
						logit.Error("Error in newStats: %e", err.Error())
					} else {
						data, err := json.Marshal(stats.data)
						if err != nil {
							logit.Error("Error marshalling data: %e", err.Error())
						}

						sendToBulker(data)
					}
				}
			}
		}
		return true, llen, queueLength, nil
	}

	return false, 0, queueLength, nil
}

//SetupRedis setup redis
func SetupRedis() {
	if os.Getenv("PLATFORM_ENV") != "test" {
		logit.Info("Redis URL: %s\n", redisURL())
	}
	pool = newPool(redisURL())
	redisPassword = redisPW()
}

func redisURL() string {
	// If flag is set, use that.
	if redisPtr != "" {
		return redisPtr
	}
	// Next use an environment variable, if it's set
	if value, ok := os.LookupEnv("PLS_REDIS_URL"); ok {
		return value
	}
	// Lastly return default
	return "127.0.0.1:6379"
}

func redisPW() string {
	// If flag is set, use that.
	if redisPwPtr != "" {
		return redisPwPtr
	}
	// Next use an environment variable, if it's set
	if value, ok := os.LookupEnv("PLS_REDIS_PASSWORD"); ok {
		return value
	}
	// Lastly return default
	return ""
}

func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		MaxActive:   0,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				logit.Error("Error on connecting to Redis: %e", err.Error())
				return nil, err
			}
			if redisPassword != "" {
				if _, err = c.Do("AUTH", redisPassword); err != nil {
					logit.Error("Error on authenticating to Redis: %e", err.Error())
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func startRedisSingle(redisKey string) {
	for {
		query, err := getLog(redisKey)
		if err != nil {
			logit.Error("Error getting log from Redis: %e", err.Error())
		}
		if query != nil {
			addToQueries(currentMinute(), query)
		} else {
			logit.Info("No new queries found. Waiting 5 seconds.")
			time.Sleep((time.Second * 5))
		}
	}
}

func getDuration(nap int) time.Duration {
	sleepDuration := time.Duration(math.Ceil((float64(nap) + 0.01) / 4.0))
	if sleepDuration > 60 {
		sleepDuration = 60
	}
	return sleepDuration
}
