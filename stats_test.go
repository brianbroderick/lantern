package main

import (
	"os"
	"testing"

	logit "github.com/brianbroderick/logit"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

// TestStats is basically an end to end integration test for a stats key
func TestStats(t *testing.T) {
	initialSetup()
	SetupElastic()
	truncateElasticSearch()

	conn := pool.Get()
	defer conn.Close()

	key := "stats"

	sample := readPayload("stats.json")
	conn.Do("LPUSH", key, sample)

	llen, err := conn.Do("LLEN", key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), llen)

	err = bulkProc["bulk"].Flush()
	if err != nil {
		logit.Error("Error flushing messages: %e", err.Error())
	}

	conn.Do("DEL", key)
	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
}
