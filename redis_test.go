package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestToAndFromRedis(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("execute.json")
	conn.Do("LPUSH", redisKey(), sample)

	llen, err := conn.Do("LLEN", redisKey())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), llen)

	message := "duration: 0.051 ms  execute <unnamed>: select * from servers"
	query, err := getLog()
	assert.NoError(t, err)
	assert.Equal(t, message, query.message)

	assert.Equal(t, 0.051, query.duration)
	assert.Equal(t, "execute", query.commandTag)
	assert.Equal(t, "<unnamed>", query.prepared)
	assert.Equal(t, "select * from servers", query.query)

	conn.Do("DEL", redisKey())
}

func readPayload(filename string) []byte {
	dat, err := ioutil.ReadFile("./sample_payloads/" + filename)
	check(err)
	return dat
}
