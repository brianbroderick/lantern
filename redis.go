package main

import (
	"encoding/json"
	"math"
	"os"
	"time"

	logit "github.com/brettallred/go-logit"
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

		query, err := newQuery(data, redisKey)
		if err != nil {
			return nil, err
		}
		return query, nil
	}
	return nil, nil
}

func startRedisBatch(redisKey string) {
	nap := 0
	lastLog := 0
	var sleepDuration time.Duration

	for {
		ok, err := getMultiLog(redisKey)
		if err != nil {
			logit.Error(" Error in getMultiLog: %e", err.Error())
		}
		if ok == false {
			sleepDuration = time.Duration(math.Ceil((float64(nap) + 0.01) / 10.0))
			if sleepDuration > 15 {
				sleepDuration = 15
			}

			nap += int(sleepDuration)
			time.Sleep((time.Second * sleepDuration))

			if nap-lastLog >= 20 {
				logit.Info(" Seconds since last Redis log received from %s key: %d", redisKey, nap)
				lastLog = nap
			}
		} else {
			nap = 0
			lastLog = 0
		}
	}
}

func getMultiLog(redisKey string) (bool, error) {
	conn := pool.Get()
	defer conn.Close()

	// get list length
	l, err := conn.Do("LLEN", redisKey)
	if err != nil {
		return true, err
	}

	llen, err := redis.Int64(l, err)
	if err != nil {
		return true, err
	}

	// LPOP at most 50 at a time
	if llen > 50 {
		llen = 50
	}

	if llen > 0 {
		conn.Send("MULTI")
		var n int64
		for n = 0; n < llen; n++ {
			conn.Send("LPOP", redisKey)
		}

		reply, err := redis.Values(conn.Do("EXEC"))
		if err != nil {
			return true, err
		}

		for _, datum := range reply {
			if datum != nil {
				q, err := redis.Bytes(datum, err)
				if err != nil {
					return true, err
				}

				query, err := newQuery(q, redisKey)
				if err != nil {
					logit.Error(" Error in newQuery: %e", err.Error())
				} else {
					addToQueries(roundToMinute(query.timestamp), query)
				}
			}
		}
		return true, nil
	}

	return false, nil
}

//SetupRedis setup redis
func SetupRedis() {
	pool = newPool(os.Getenv("PLS_REDIS_URL"))
	redisPassword = os.Getenv("PLS_REDIS_PASSWORD")
}

func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				logit.Error(" Error on connecting to Redis: %e", err.Error())
				return nil, err
			}
			if redisPassword != "" {
				if _, err = c.Do("AUTH", redisPassword); err != nil {
					logit.Error(" Error on authenticating to Redis: %e", err.Error())
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
			logit.Error(" Error getting log from Redis: %e", err.Error())
		}
		if query != nil {
			addToQueries(currentMinute(), query)
		} else {
			logit.Info(" No new queries found. Waiting 5 seconds.")
			time.Sleep((time.Second * 5))
		}
	}
}
