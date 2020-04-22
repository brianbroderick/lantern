package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/brianbroderick/logit"
	"github.com/stretchr/testify/assert"
	elastic "gopkg.in/olivere/elastic.v6"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func truncateElasticSearch() {
	for _, index := range indices() {
		_, error := clients["bulk"].DeleteIndex(index).Do(context.Background())
		if error != nil {
			panic(error)
		}
	}
	putTemplate(clients["bulk"])
}

func TestBasicFlow(t *testing.T) {
	initialSetup()
	SetupElastic()
	truncateElasticSearch()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("execute_without_comment.json")
	conn.Do("LPUSH", redisKey(), sample)

	llen, err := conn.Do("LLEN", redisKey())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), llen)

	message := "duration: 0.051 ms  execute <unnamed>: select * from servers where id IN ('1', '2', '3') and name = 'localhost'"

	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, message, query.message)

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

func TestElixirFlow(t *testing.T) {
	initialSetup()
	SetupElastic()
	truncateElasticSearch()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("elixir.json")
	conn.Do("LPUSH", redisKey(), sample)

	llen, err := conn.Do("LLEN", redisKey())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), llen)

	message := "duration: 17.646 ms execute N/A: INSERT INTO \"raw\".\"raw_events\" (\"data\",\"published_at\",\"inserted_at\",\"updated_at\") VALUES ($1,$2,$3,$4) RETURNING \"id\""
	comments := []string(nil)
	query, err := getLog(redisKey())
	assert.NoError(t, err)
	assert.Equal(t, message, query.message)
	assert.Equal(t, comments, query.comments)

	assert.Equal(t, 17.646, query.totalDuration)
	assert.Equal(t, "execute", query.preparedStep)
	assert.Equal(t, "N/A", query.prepared)
	assert.Equal(t, "INSERT INTO \"raw\".\"raw_events\" (\"data\",\"published_at\",\"inserted_at\",\"updated_at\") VALUES ($1,$2,$3,$4) RETURNING \"id\"", query.query)

	pgQuery := "insert into \"raw\".\"raw_events\" (\"data\",\"published_at\",\"inserted_at\",\"updated_at\") values ($1,$2,$3,$4) returning \"id\""
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

	totalDuration := getRecord(t, 1000, "elixir_user")
	assert.Equal(t, 35.292, totalDuration)

	conn.Do("DEL", redisKey())
	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
}

func TestBadPayload(t *testing.T) {
	initialSetup()
	SetupElastic()
	truncateElasticSearch()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("nested_payload.json")
	conn.Do("LPUSH", redisKey(), sample)

	llen, err := conn.Do("LLEN", redisKey())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), llen)

	query, err := getLog(redisKey())
	addToQueries(mockCurrentMinute(), query)

	iterOverQueries()
	assert.Equal(t, 0, len(batchMap))

	err = bulkProc["bulk"].Flush()
	if err != nil {
		logit.Error("Error flushing messages: %e", err.Error())
	}

	// There is no record, on ES because an array of different types can not be published
	// to elastic search. We need to log out these failed publish attempts.
	//getRecord(t,1000 )

	// We should see logs from the afterBulkCommit function

	conn.Do("DEL", redisKey())
	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()
}

func getRecord(t *testing.T, wait time.Duration, username string) float64 {
	// fmt.Printf("getR: %s \n", indexName())

	termQuery := elastic.NewTermQuery("user_name", username)
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
		fmt.Printf("Found no records, waiting %d ms...\n", wait)
		time.Sleep(wait * time.Millisecond)
		if wait+wait > 4000 {
			t.Fatalf("Max timeout while attmpting to get elastic search results.")
		}
		return getRecord(t, wait+wait, username)
	}
	return -1.0
}

// func TestColor(t *testing.T) {
// color.Set(color.FgYellow)
// logit.Info("Sent %d messages to ES Bulk Processor", 72)
// color.Unset()

// fmt.Printf("This is a %s and this is %s.\n", yellow("warning"), red("error"))

// fmt.Printf("This is a %s and this is %s.\n", yellow("warning"), red("error"))
// 	logit.Info("%s messages processed from %s since last reset", yellow(2), green("blah"))
// 	logit.Info("Current queue length for %s is %s", green("blah"), red(6))
// }

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
