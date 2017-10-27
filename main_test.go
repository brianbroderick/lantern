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
	query, err := getLog()
	assert.NoError(t, err)
	assert.Equal(t, message, query.message)

	assert.Equal(t, 0.051, query.duration)
	assert.Equal(t, "execute", query.preparedStep)
	assert.Equal(t, "<unnamed>", query.prepared)
	assert.Equal(t, "select * from servers where id IN ('1', '2', '3') and name = 'localhost'", query.query)

	pgQuery := "select * from servers where id IN (?) and name = ?"
	assert.Equal(t, pgQuery, query.normalizedQuery)

	assert.Equal(t, 0, len(batchMap))
	_, ok := batchMap[batch{mockCurrentMinute(), query.uniqueSha}]
	assert.False(t, ok)
	addToQueries(mockCurrentMinute(), query)
	assert.Equal(t, 1, len(batchMap))
	assert.Equal(t, int32(1), batchMap[batch{mockCurrentMinute(), query.uniqueSha}].count)

	addToQueries(mockCurrentMinute(), query)
	_, ok = batchMap[batch{mockCurrentMinute(), query.uniqueSha}]
	assert.True(t, ok)
	assert.Equal(t, 1, len(batchMap))
	assert.Equal(t, int32(2), batchMap[batch{mockCurrentMinute(), query.uniqueSha}].count)

	iterOverQueries()
	assert.Equal(t, 0, len(batchMap))

	err = bulkProc["bulk"].Flush()
	if err != nil {
		logit.Error("Error flushing messages: %e", err.Error())
	}
	getRecord()

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

func mockCurrentMinute() time.Time {
	d := time.Date(2017, time.October, 27, 19, 57, 5, 1250, time.UTC)
	return d.UTC().Round(time.Minute)
}

func getRecord() {
	termQuery := elastic.NewTermQuery("normalized_sha", "9ac8616b76d626c6b06372f9834cce48f7660c3a")
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
		fmt.Printf("Found a total of %d tweets\n", result.Hits.TotalHits)

		for _, hit := range result.Hits.Hits {
			// hit.Index contains the name of the index

			var data map[string]*json.RawMessage
			if err := json.Unmarshal(*hit.Source, &data); err != nil {
				logit.Error("Error unmarshalling data: %e", err.Error())
			}

			var duration float64
			if source, pres := data["duration"]; pres {
				if err := json.Unmarshal(*source, &duration); err != nil {
					logit.Error("Error unmarshalling duration: %e", err.Error())
				}
			}

			fmt.Printf("First record found has a duration of %f\n", duration)
		}
	} else {
		// No hits
		fmt.Print("Found no tweets, waiting 500ms...\n")
		time.Sleep(500 * time.Millisecond)
		getRecord()
	}
}
