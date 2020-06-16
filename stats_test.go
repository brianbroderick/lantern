package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	logit "github.com/brianbroderick/logit"
	elastic "gopkg.in/olivere/elastic.v6"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

// TestStats is basically an end to end integration test for a stats key
// func TestStats(t *testing.T) {
// 	fmt.Println("TestStats")
// 	initialSetup()
// 	SetupElastic()
// 	truncateElasticSearch()

// 	conn := pool.Get()
// 	defer conn.Close()

// 	key := "stats"

// 	sample := readPayload("stats.json")
// 	conn.Do("LPUSH", key, sample)

// 	llen, err := conn.Do("LLEN", key)
// 	assert.NoError(t, err)
// 	assert.Equal(t, int64(1), llen)

// 	err = bulkProc["bulk"].Flush()
// 	if err != nil {
// 		logit.Error("Error flushing messages: %e", err.Error())
// 	}

// 	totalDuration := getStatsRecord(t, 1000, "myapp")
// 	assert.Equal(t, 0.102, totalDuration)

// 	conn.Do("DEL", key)
// 	defer bulkProc["bulk"].Close()
// 	defer clients["bulk"].Stop()
// }

func getStatsRecord(t *testing.T, wait time.Duration, app string) int64 {
	// fmt.Printf("getR: %s \n", indexName())

	termQuery := elastic.NewTermQuery("application_name", app)
	result, err := clients["bulk"].Search().
		Index(statsIndexName()).
		Type("pgstats").
		Query(termQuery).
		From(0).Size(1).
		Do(context.Background())
	if err != nil {
		panic(err)
	}

	if result.Hits.TotalHits > 0 {
		fmt.Printf("Found a total of %d stat record(s)\n", result.Hits.TotalHits)

		for _, hit := range result.Hits.Hits {
			// hit.Index contains the name of the index

			var data map[string]*json.RawMessage
			if err := json.Unmarshal(*hit.Source, &data); err != nil {
				logit.Error("Error unmarshalling data: %e", err.Error())
			}

			var failedEvents int64
			if source, pres := data["failed_events"]; pres {
				if err := json.Unmarshal(*source, &failedEvents); err != nil {
					logit.Error("Error unmarshalling failedEvents: %e", err.Error())
				}
			}

			fmt.Printf("First record found returned %d for the failed attempts key\n", failedEvents)
			return failedEvents
		}
	} else {
		// No hits
		fmt.Printf("Found no records, waiting %d ms...\n", wait)
		time.Sleep(wait * time.Millisecond)
		if wait+wait > 4000 {
			t.Fatalf("Max timeout while attmpting to get elastic search results.")
		}
		return getStatsRecord(t, wait+wait, app)
	}
	return -1.0
}
