package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestRegexMessage(t *testing.T) {
	// check standard statement
	message := "duration: 0.083 ms  statement: SET time zone 'UTC'"
	result := regexMessage(message)
	assert.Equal(t, "statement", result["preparedStep"])

	// check prepared statement
	message = "duration: 0.066 ms  bind <unnamed>: select * from servers where id = 1"
	result = regexMessage(message)
	assert.Equal(t, "bind", result["preparedStep"])

	// check non-greedy to colon
	message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah'"
	result = regexMessage(message)
	assert.Equal(t, "bind", result["preparedStep"])

	// check multiline
	message = `duration: 0.066 ms  bind <unnamed>: select * from servers
	where name = 'blah:blah'`
	multiLine := `select * from servers
	where name = 'blah:blah'`
	result = regexMessage(message)
	assert.Equal(t, multiLine, result["grokQuery"])
}

func TestComments(t *testing.T) {
	var q = new(query)
	// No comment
	q.message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah'"
	assert.NotPanics(t, func() { extractComments(q) })

	// Legit comment
	q.message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah' /*application:Rails,controller:users,action:search,line:monitor.rb:214:in '::Weekly::Digest/mon_synchronize'*/"
	assert.NotPanics(t, func() { extractComments(q) })

	// not complete comment
	q.message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah' /*application:Rails*/"
	assert.NotPanics(t, func() { extractComments(q) })

	// Empty comment
	q.message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah' /**/"
	assert.NotPanics(t, func() { extractComments(q) })

	// illegit comment with colon
	q.message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah' /*ask:yourmom*/"
	assert.NotPanics(t, func() { extractComments(q) })

	q.message = "/*ask:you:*/"
	assert.NotPanics(t, func() { extractComments(q) })

	q.message = "/*:*/"
	assert.NotPanics(t, func() { extractComments(q) })

	q.message = `duration: 0.066 ms  bind <unnamed>: select * from multiline /*
	: blah */`
	assert.NotPanics(t, func() { extractComments(q) })

	// Multiples
	q.message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah' /*application:Rails*/ /*application:Sidekiq*/"
	assert.NotPanics(t, func() { extractComments(q) })
}

func TestFlowWithoutComment(t *testing.T) {
	initialSetup()
	// truncateElasticSearch()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("execute_without_comment.json")
	conn.Do("LPUSH", redisKey(), sample)

	llen, err := conn.Do("LLEN", redisKey())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), llen)

	message := "duration: 0.051 ms  execute <unnamed>: select * from servers where id IN ('1', '2', '3') and name = 'localhost'"
	comments := []string(nil)
	query, err := getLog(redisKey())

	assert.NoError(t, err)
	assert.Equal(t, message, query.message)
	assert.Equal(t, comments, query.comments)
}

func TestUpdateWaiting(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("update_waiting.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "process 11451 acquired ExclusiveLock on page 0 of relation 519373 of database 267504 after 1634.121 ms"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	message := "update some_table set mms_url = $1, sms_text = $2, message_sid = $3, updated_at = $4 where some_table.id = $5"
	assert.Equal(t, message, query.uniqueStr)
}

func TestFatal(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("fatal.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "some fatal error"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	message := "1234"
	assert.Equal(t, message, query.uniqueStr)
}

func TestConnReceived(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("connection_received.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "connection received: host=10.0.1.168 port=38634"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	uniqueStr := "connection_received"
	assert.Equal(t, uniqueStr, query.uniqueStr)
}

func TestDisconnection(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("disconnection.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "disconnection: session time: 0:00:00.074 user=root database= host=10.0.1.168 port=56544"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	uniqueStr := "disconnection"
	assert.Equal(t, uniqueStr, query.uniqueStr)
}

func TestConnRepl(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("repl_connection.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "replication connection authorized: user=root SSL enabled (protocol=TLSv1.2, cipher=ECDHE-RSA-AES256-GCM-SHA384, compression=off)"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	uniqueStr := "connection_replication"
	assert.Equal(t, uniqueStr, query.uniqueStr)
}

func TestCheckpointStarting(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("checkpoint_starting.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "checkpoint starting: time"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	uniqueStr := "checkpoint_starting"
	assert.Equal(t, uniqueStr, query.uniqueStr)
}

func TestCheckpointComplete(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("checkpoint_complete.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "checkpoint complete: time"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	uniqueStr := "checkpoint_complete"
	assert.Equal(t, uniqueStr, query.uniqueStr)
}

func TestVacuum(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("vacuum.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "automatic vacuum of table \"app.public.some_table\": blah blah"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	uniqueStr := "vacuum_table app.public.some_table"
	assert.Equal(t, uniqueStr, query.uniqueStr)
}

func TestAnalyze(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("analyze.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "automatic analyze of table \"app.public.some_table\": system usage: CPU 0.00s/0.02u sec elapsed 0.15 sec"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	uniqueStr := "analyze_table app.public.some_table"
	assert.Equal(t, uniqueStr, query.uniqueStr)
}

func TestConnectionReset(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("connection_reset.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "could not receive data from client: Connection reset by peer"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	uniqueStr := "connection_reset"
	assert.Equal(t, uniqueStr, query.uniqueStr)
}

func TestTempTable(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("temp_table.json")
	conn.Do("LPUSH", redisKey(), sample)

	message := "temporary file: path \"base/pgsql_tmp/pgsql_tmp73093.7\", size 2576060"
	grokQuery := "SELECT DISTINCT users.* FROM users WHERE users.active = 't'"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, message, query.message)
	assert.Equal(t, grokQuery, query.query)
	assert.Equal(t, int64(2576060), query.tempTable)

	conn.Do("DEL", redisKey())
}
