package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestLogFlow(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("execute.json")
	conn.Do("LPUSH", redisKey(), sample)

	llen, err := conn.Do("LLEN", redisKey())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), llen)

	message := "duration: 0.051 ms  execute <unnamed>: select * from servers where id IN ('1', '2', '3') and name = 'localhost'"
	query, err := getLog()
	assert.NoError(t, err)
	assert.Equal(t, message, query.message)

	assert.Equal(t, 0.051, query.duration)
	assert.Equal(t, "execute", query.preparedStep)
	assert.Equal(t, "<unnamed>", query.prepared)
	assert.Equal(t, "select * from servers where id IN ('1', '2', '3') and name = 'localhost'", query.query)

	pgQuery := "select * from servers where id IN (?) and name = ?"
	assert.Equal(t, pgQuery, query.normalizedQuery)

	data, err := json.Marshal(query.data)
	assert.NoError(t, err)
	saveToElastic(data)

	// fmt.Println("")
	// fmt.Println(query.uniqueSha)

	_, ok := batchMap[batch{mockCurrentMinute(), query.uniqueSha}]
	assert.False(t, ok)
	addToQueries(mockCurrentMinute(), query)

	assert.Equal(t, int32(1), batchMap[batch{mockCurrentMinute(), query.uniqueSha}].count)

	addToQueries(mockCurrentMinute(), query)
	_, ok = batchMap[batch{mockCurrentMinute(), query.uniqueSha}]
	assert.True(t, ok)
	assert.Equal(t, int32(2), batchMap[batch{mockCurrentMinute(), query.uniqueSha}].count)

	conn.Do("DEL", redisKey())
	defer clients["bulk"].Stop()
}

func readPayload(filename string) []byte {
	dat, err := ioutil.ReadFile("./sample_payloads/" + filename)
	check(err)
	return dat
}

func mockCurrentMinute() time.Time {
	d := time.Date(2017, time.November, 10, 23, 19, 5, 1250, time.UTC)
	return d.UTC().Round(time.Minute)
}
