package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestToAndFromRedis(t *testing.T) {
	initialSetup()

	sample := readPayload("execute.json")
	conn := pool.Get()
	defer conn.Close()
	conn.Do("LPUSH", redisKey(), sample)

	llen, err := conn.Do("LLEN", redisKey())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), llen)

	var data json.RawMessage

	// redis, err := getLog()
	reply, err := conn.Do("LPOP", redisKey())
	assert.NoError(t, err)

	data, err = redis.Bytes(reply, err)
	assert.NoError(t, err)

	query, err := newQuery(data)
	assert.NoError(t, err)

	fmt.Printf("%+v\n", query)

	conn.Do("DEL", redisKey())
}

func readPayload(filename string) []byte {
	dat, err := ioutil.ReadFile("./sample_payloads/" + filename)
	check(err)
	return dat
}

// func getLog() (json.RawMessage, error) {
// 	var data json.RawMessage

// 	conn := pool.Get()
// 	defer conn.Close()

// 	log, err := redis.Values(conn.Do("LPOP", "postgres"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	if _, err := redis.Scan(log, &data); err != nil {
// 		return nil, err
// 	}

// 	return data, nil
// }
