package main

import (
	"os"
	"testing"
	"time"

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

			q, suppressed, err := newQuery(d, redisQueues[0])
			assert.NoError(t, err)

			if suppressedCommandTag[q.commandTag] {
				assert.Equal(t, suppressed, true)
			} else {
				assert.Equal(t, suppressed, false)
			}
		}
	}

	llen, err = conn.Do("LLEN", pipeline)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), llen)

	conn.Do("DEL", pipeline)
}

func TestGetDuration(t *testing.T) {
	assert.Equal(t, time.Duration(1), getDuration(0))
	assert.Equal(t, time.Duration(9), getDuration(35))
	assert.Equal(t, time.Duration(60), getDuration(300))

	nap := 0
	var sleepDuration time.Duration

	// 1 second each
	for i := 0; i < 4; i++ {
		sleepDuration = getDuration(nap)
		nap += int(sleepDuration)
	}
	assert.Equal(t, 4, nap)

	// Starting with 4, then +2, +2, +3, +3
	for i := 0; i < 4; i++ {
		sleepDuration = getDuration(nap)
		nap += int(sleepDuration)
	}
	assert.Equal(t, 14, nap)

	// Starting with 14, then +4, +5, +6, +8
	for i := 0; i < 4; i++ {
		sleepDuration = getDuration(nap)
		nap += int(sleepDuration)
	}
	assert.Equal(t, 37, nap)
}
