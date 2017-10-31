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

func startRedis() {
	for {
		query, err := getLog()
		if err != nil {
			logit.Error(" Error getting log from Redis: %e", err.Error())
		}
		if query != nil {
			addToQueries(currentMinute(), query)
		} else {
			logit.Info(" No queries found. Waiting 5 seconds.")
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

//SetupRedis setup redis
func SetupRedis() {
	pool = newPool(os.Getenv("REDIS_URL"))
	redisPassword = os.Getenv("REDIS_PASSWORD")
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
