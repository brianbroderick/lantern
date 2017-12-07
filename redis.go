package main

import (
	"encoding/json"
	"os"
	"time"

	logit "github.com/brettallred/go-logit"
	"github.com/garyburd/redigo/redis"
)

var (
	pool          *redis.Pool
	redisPassword string
)

func startRedisSingle() {
	for {
		query, err := getLog()
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

func getLog() (*query, error) {
	var data json.RawMessage

	conn := pool.Get()
	defer conn.Close()

	reply, err := conn.Do("LPOP", redisKey())
	if err != nil {
		return nil, err
	}

	if reply != nil {
		data, err = redis.Bytes(reply, err)
		if err != nil {
			return nil, err
		}

		query, err := newQuery(data)
		if err != nil {
			return nil, err
		}
		return query, nil
	}
	return nil, nil
}

func startRedisBatch() {
	nap := 0

	for {
		ok, err := getMultiLog()
		if err != nil {
			logit.Error(" Error in getMultiLog: %e", err.Error())
		}
		if ok == false {
			nap++
			time.Sleep((time.Second * 1))
			if nap%10 == 0 {
				logit.Info(" Seconds since last Redis log received: %d", nap)
			}
		} else {
			nap = 0
		}
	}
}

func getMultiLog() (bool, error) {
	conn := pool.Get()
	defer conn.Close()

	// get list length
	l, err := conn.Do("LLEN", redisKey())
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
			conn.Send("LPOP", redisKey())
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

				query, err := newQuery(q)
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
