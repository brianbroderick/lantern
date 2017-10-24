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

	// data, err := json.Marshal(query.data)
	// saveToElastic(data)
	// assert.NoError(t, err)

	// fmt.Println("")
	// fmt.Println(query.uniqueSha)

	_, ok := queryMap[query.uniqueSha]
	assert.False(t, ok)

	addToQueries(query)
	assert.Equal(t, int32(1), queryMap[query.uniqueSha].count)

	addToQueries(query)

	_, ok = queryMap[query.uniqueSha]
	assert.True(t, ok)
	assert.Equal(t, int32(2), queryMap[query.uniqueSha].count)

	// que, ok := queryMap[query.uniqueSha]
	// assert.False(t, ok)
	// fmt.Printf("%v+", que)

	// assert.Equal(t, 1, queryMap["9ac8616b76d626c6b06372f9834cce48f7660c3a"].count)

	conn.Do("DEL", redisKey())
}

func readPayload(filename string) []byte {
	dat, err := ioutil.ReadFile("./sample_payloads/" + filename)
	check(err)
	return dat
}
