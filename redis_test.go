package main

import (
	"os"
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestPipeline(t *testing.T) {
	pipeline := "pipeline"
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("execute.json")
	conn.Do("LPUSH", pipeline, sample)

	sample = readPayload("bind.json")
	conn.Do("LPUSH", pipeline, sample)

	llen, err := conn.Do("LLEN", pipeline)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), llen)

	conn.Send("MULTI")
	conn.Send("LPOP", pipeline)
	conn.Send("LPOP", pipeline)
	conn.Send("LPOP", pipeline)
	conn.Send("LPOP", pipeline)
	reply, err := redis.Values(conn.Do("EXEC"))
	assert.NoError(t, err)

	for _, datum := range reply {
		if datum != nil {
			d, err := redis.Bytes(datum, err)
			assert.NoError(t, err)

			_, err = newQuery(d)
			assert.NoError(t, err)
		}
	}

	llen, err = conn.Do("LLEN", pipeline)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), llen)

	conn.Do("DEL", pipeline)
}

// func getMultiLog(batchSize int) (*query, error) {
// 	var data json.RawMessage

// 	conn := pool.Get()
// 	defer conn.Close()

// 	reply, err := conn.Do("LPOP", redisKey())
// 	if err != nil {
// 		return nil, err
// 	}

// 	data, err = redis.Bytes(reply, err)
// 	if err != nil {
// 		return nil, err
// 	}

// 	query, err := newQuery(data)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return query, nil
// }
