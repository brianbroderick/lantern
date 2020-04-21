package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
	"time"

	logit "github.com/brianbroderick/logit"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

// TestFlow is basically an end to end integration test
func TestFlow(t *testing.T) {
	initialSetup()
	SetupElastic()
	truncateElasticSearch()

	conn := pool.Get()
	defer conn.Close()

	// Test if code_source is nested
	bp := readPayload("nested_payload.json")
	conn.Do("LPUSH", redisKey(), bp)

	sample := readPayload("execute.json")
	conn.Do("LPUSH", redisKey(), sample)

	llen, err := conn.Do("LLEN", redisKey())
	assert.NoError(t, err)
	assert.Equal(t, int64(2), llen)

	message := "duration: 0.051 ms  execute <unnamed>: select * from servers where id IN ('1', '2', '3') and name = 'localhost'"
	comments := []string{"/*application:Rails,controller:users,action:search,line:/usr/local/rvm/rubies/ruby-2.3.7/lib/ruby/2.3.0/monitor.rb:214:in mon_synchronize*/"}
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, message, query.message)
	assert.Equal(t, comments, query.comments)

	assert.Equal(t, 0.051, query.totalDuration)
	assert.Equal(t, "execute", query.preparedStep)
	assert.Equal(t, "<unnamed>", query.prepared)
	assert.Equal(t, "select * from servers where id IN ('1', '2', '3') and name = 'localhost'", query.query)

	pgQuery := "select * from servers where id in ($1, $2, $3) and name = $4"
	assert.Equal(t, pgQuery, query.uniqueStr)

	assert.Equal(t, 0, len(batchMap))
	_, ok := batchMap[batch{mockCurrentMinute(), query.uniqueSha}]
	assert.False(t, ok)
	addToQueries(mockCurrentMinute(), query)
	assert.Equal(t, 1, len(batchMap))
	assert.Equal(t, int32(1), batchMap[batch{mockCurrentMinute(), query.uniqueSha}].totalCount)

	addToQueries(mockCurrentMinute(), query)
	_, ok = batchMap[batch{mockCurrentMinute(), query.uniqueSha}]
	assert.True(t, ok)
	assert.Equal(t, 1, len(batchMap))
	assert.Equal(t, int32(2), batchMap[batch{mockCurrentMinute(), query.uniqueSha}].totalCount)

	iterOverQueries()
	assert.Equal(t, 0, len(batchMap))

	err = bulkProc["bulk"].Flush()
	if err != nil {
		logit.Error("Error flushing messages: %e", err.Error())
	}

	totalDuration := getRecord(t, 1000, "execute_user")
	assert.Equal(t, 0.102, totalDuration)

	conn.Do("DEL", redisKey())
	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
}

func readPayload(filename string) []byte {
	dat, err := ioutil.ReadFile("./sample_payloads/" + filename)
	check(err)
	return dat
}

// TestCurrentMinute basically tests currentMinute()
func TestCurrentMinute(t *testing.T) {
	d := time.Date(2017, time.November, 10, 23, 19, 5, 1250, time.UTC)
	minute := d.UTC().Round(time.Minute)
	assert.Equal(t, 0, minute.Second())
}

func TestRound(t *testing.T) {
	r := round(0.564627465465, 0.5, 5)
	assert.Equal(t, 0.56463, r)
}

func mockCurrentMinute() time.Time {
	d := time.Date(2017, time.October, 27, 19, 57, 5, 1250, time.UTC)
	return d.UTC().Round(time.Minute)
}

func TestPopulateRedisQueues(t *testing.T) {
	populateRedisQueues("test")
	expected := []string{"test"}
	assert.Equal(t, expected, redisQueues)

	populateRedisQueues("bob,bill,jane")
	expected = []string{"bob", "bill", "jane"}
	assert.Equal(t, expected, redisQueues)

	populateRedisQueues(" ben,   jimmy, greg   ")
	expected = []string{"ben", "jimmy", "greg"}
	assert.Equal(t, expected, redisQueues)
}

func TestParseTime(t *testing.T) {
	tempTime := "2014-10-16T15:50:35.840+0000"
	parsedTime, err := time.Parse(longForm, tempTime)
	if err != nil {
		fmt.Printf("%e \n\n", err)
	}
	expected := "2014-10-16 15:50:35.84 +0000 UTC"
	actual := fmt.Sprint(parsedTime)
	assert.Equal(t, expected, actual)
}

func TestRoundToMinute(t *testing.T) {
	tempTime := "2011-10-16T15:50:15.840+0000"
	parsedTime, err := time.Parse(longForm, tempTime)
	if err != nil {
		fmt.Printf("%e \n\n", err)
	}

	rounded := roundToMinute(parsedTime)
	expected := "2011-10-16 15:50:00 +0000 UTC"
	assert.Equal(t, expected, rounded.String())
}

func TestFindAllStringSubmatch(t *testing.T) {
	str := "hello /* foobar */ /* blahboo */"
	r := regexp.MustCompile(`(/\*.*?:.*?\*/)`)
	match := r.FindAllStringSubmatch(str, -1)

	assert.Equal(t, 0, len(match))
}
