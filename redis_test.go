package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestToAndFromRedis(t *testing.T) {
	initialSetup()

	sample := readPayload("execute.json")
	conn := pool.Get()
	defer conn.Close()

	conn.Do("LPUSH", RedisKey, sample)

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
