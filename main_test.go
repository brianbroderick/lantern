package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"

	logit "github.com/brettallred/go-logit"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

// TestFlow is basically an end to end integration test
func TestFlow(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("execute.json")
	conn.Do("LPUSH", redisKey(), sample)

	llen, err := conn.Do("LLEN", redisKey())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), llen)

	message := "duration: 0.051 ms  execute <unnamed>: select * from servers where id IN ('1', '2', '3') and name = 'localhost'"
	comments := []string{"/*application:Rails,controller:users,action:search,line:/usr/local/rvm/rubies/ruby-2.3.7/lib/ruby/2.3.0/monitor.rb:214:in `mon_synchronize'*/"}
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, message, query.message)
	assert.Equal(t, comments, query.comments)
	assert.Equal(t, "Rails", query.mvcApplication)
	assert.Equal(t, "users", query.mvcController)
	assert.Equal(t, "search", query.mvcAction)
	assert.Equal(t, "/usr/local/rvm/rubies/ruby-2.3.7/lib/ruby/2.3.0/monitor.rb:214:in `mon_synchronize'", query.mvcCodeLine)

	assert.Equal(t, 0.051, query.totalDuration)
	assert.Equal(t, "execute", query.preparedStep)
	assert.Equal(t, "<unnamed>", query.prepared)
	assert.Equal(t, "select * from servers where id IN ('1', '2', '3') and name = 'localhost'", query.query)

	pgQuery := "select * from servers where id in (?) and name = ?"
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
	totalDuration := getRecord()
	assert.Equal(t, 0.102, totalDuration)

	conn.Do("DEL", redisKey())
	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
}

func TestTempTable(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("temp_table.json")
	conn.Do("LPUSH", redisKey(), sample)

	message := "temporary file: path \"base/pgsql_tmp/pgsql_tmp73093.7\", size 2576060"
	grokQuery := "SELECT DISTINCT \"users\".* FROM \"users\" LEFT JOIN location_users ON location_users.employee_id = users.id WHERE \"users\".\"active\" = 't' AND (location_users.location_id = 17511 OR (users.organization_id = 7528 AND users.role = 'Client'))"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, message, query.message)
	assert.Equal(t, grokQuery, query.query)
	assert.Equal(t, int64(2576060), query.tempTable)

	conn.Do("DEL", redisKey())
	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
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

	message := "update \"review_invitations\" set \"mms_url\" = $1, \"sms_text\" = $2, \"message_sid\" = $3, \"updated_at\" = $4 where \"review_invitations\".\"id\" = $5"
	assert.Equal(t, message, query.uniqueStr)

	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
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

	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
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

	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
}

func TestDisconnection(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("disconnection.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "disconnection: session time: 0:00:00.074 user=q55cd17435 database= host=10.0.1.168 port=56544"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	uniqueStr := "disconnection"
	assert.Equal(t, uniqueStr, query.uniqueStr)

	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
}

func TestConnRepl(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("repl_connection.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "replication connection authorized: user=q55cd17435 SSL enabled (protocol=TLSv1.2, cipher=ECDHE-RSA-AES256-GCM-SHA384, compression=off)"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	uniqueStr := "connection_replication"
	assert.Equal(t, uniqueStr, query.uniqueStr)

	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
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

	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
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

	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
}

func TestVacuum(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("vacuum.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "automatic vacuum of table \"app.public.api_clients\": blah blah"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	uniqueStr := "vacuum_table app.public.api_clients"
	assert.Equal(t, uniqueStr, query.uniqueStr)

	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
}

func TestAnalyze(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("analyze.json")
	conn.Do("LPUSH", redisKey(), sample)

	notes := "automatic analyze of table \"app.public.api_clients\": system usage: CPU 0.00s/0.02u sec elapsed 0.15 sec"
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, notes, query.notes)

	uniqueStr := "analyze_table app.public.api_clients"
	assert.Equal(t, uniqueStr, query.uniqueStr)

	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
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

func getRecord() float64 {
	termQuery := elastic.NewTermQuery("user_name", "samplepayload")
	result, err := clients["bulk"].Search().
		Index(indexName()).
		Type("pglog").
		Query(termQuery).
		From(0).Size(1).
		Do(context.Background())
	if err != nil {
		panic(err)
	}

	if result.Hits.TotalHits > 0 {
		fmt.Printf("Found a total of %d record(s)\n", result.Hits.TotalHits)

		for _, hit := range result.Hits.Hits {
			// hit.Index contains the name of the index

			var data map[string]*json.RawMessage
			if err := json.Unmarshal(*hit.Source, &data); err != nil {
				logit.Error("Error unmarshalling data: %e", err.Error())
			}

			var totalDuration float64
			if source, pres := data["total_duration_ms"]; pres {
				if err := json.Unmarshal(*source, &totalDuration); err != nil {
					logit.Error("Error unmarshalling totalDuration: %e", err.Error())
				}
			}

			fmt.Printf("First record found has a total duration of %f\n", totalDuration)
			return totalDuration
		}
	} else {
		// No hits
		fmt.Print("Found no records, waiting 500ms...\n")
		time.Sleep(500 * time.Millisecond)
		return getRecord()
	}
	return -1.0
}

func getRecordWithTempTable() int64 {
	fmt.Println("getRecordWithTempTable")

	termQuery := elastic.NewTermQuery("user_name", "temp_table")
	result, err := clients["bulk"].Search().
		Index(indexName()).
		Type("pglog").
		Query(termQuery).
		From(0).Size(1).
		Do(context.Background())
	if err != nil {
		panic(err)
	}

	if result.Hits.TotalHits > 0 {
		fmt.Printf("Found a total of %d record(s)\n", result.Hits.TotalHits)

		for _, hit := range result.Hits.Hits {
			// hit.Index contains the name of the index

			var data map[string]*json.RawMessage
			if err := json.Unmarshal(*hit.Source, &data); err != nil {
				logit.Error("Error unmarshalling data: %e", err.Error())
			}

			var tempTable int64
			if source, pres := data["temp_table_size"]; pres {
				if err := json.Unmarshal(*source, &tempTable); err != nil {
					logit.Error("Error unmarshalling tempTable: %e", err.Error())
				}
			}

			fmt.Printf("First record found has a total temp table size of %d\n", tempTable)
			return tempTable
		}
	} else {
		// No hits
		fmt.Print("Found no records, waiting 500ms...\n")
		time.Sleep(500 * time.Millisecond)
		return getRecordWithTempTable()
	}
	return -1
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
