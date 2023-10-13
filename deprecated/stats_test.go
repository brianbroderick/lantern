package deprecated

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"testing"
// 	"time"

// 	logit "github.com/brianbroderick/logit"
// 	elastic "github.com/olivere/elastic/v7"
// 	"github.com/stretchr/testify/assert"
// )

// func init() {
// 	os.Setenv("PLATFORM_ENV", "test")
// }

// // TestStats is basically an end to end integration test for a stats key
// func TestStats(t *testing.T) {
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

// 	stats, err := getStat(key)
// 	assert.NoError(t, err)

// 	data, err := json.Marshal(stats.data)
// 	assert.NoError(t, err)

// 	// saveToStatsElastic(data)
// 	sendToStatsBulker(data)

// 	err = bulkProc["bulk"].Flush()
// 	if err != nil {
// 		logit.Error("Error flushing messages: %e", err.Error())
// 	}

// 	failedEvents := getStatsRecord(t, 1000, "myapp")
// 	assert.Equal(t, int64(42), failedEvents)

// 	conn.Do("DEL", key)
// }

// func getStatsRecord(t *testing.T, wait time.Duration, app string) int64 {
// 	// fmt.Printf("getR: %s \n", indexName())

// 	termQuery := elastic.NewTermQuery("application_name", app)
// 	result, err := clients["bulk"].Search().
// 		Index(statsIndexName()).
// 		Type("pgstats").
// 		Query(termQuery).
// 		From(0).Size(1).
// 		Do(context.Background())
// 	if err != nil {
// 		panic(err)
// 	}

// 	if result.Hits.TotalHits.Value > 0 {
// 		logit.Info("Found a total of %d stat record(s)\n", result.Hits.TotalHits.Value)

// 		for _, hit := range result.Hits.Hits {
// 			// hit.Index contains the name of the index

// 			var data map[string]*json.RawMessage
// 			if err := json.Unmarshal(hit.Source, &data); err != nil {
// 				logit.Error("Error unmarshalling data: %e", err.Error())
// 			}

// 			var failedEvents int64
// 			if source, pres := data["failed_events"]; pres {
// 				if err := json.Unmarshal(*source, &failedEvents); err != nil {
// 					logit.Error("Error unmarshalling failedEvents: %e", err.Error())
// 				}
// 			}

// 			return failedEvents
// 		}
// 	} else {
// 		// No hits
// 		fmt.Printf("Found no records, waiting %d ms...\n", wait)
// 		time.Sleep(wait * time.Millisecond)
// 		if wait+wait > 4000 {
// 			t.Fatalf("Max timeout while attmpting to get elastic search results.")
// 		}
// 		return getStatsRecord(t, wait+wait, app)
// 	}
// 	return -1.0
// }
