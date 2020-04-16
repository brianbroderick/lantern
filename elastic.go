package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

// SetupElastic sets up elastic conn
func SetupElastic() {
	// Coming from Docker, sleep a few seconds to make sure ES is running
	if elasticURL() == "http://elasticsearch:9200" {
		log.Printf("Using docker, waiting for ES to spin up")
		time.Sleep(10 * time.Second)
	}
	client, err := elastic.NewClient(elastic.SetURL(elasticURL()), elastic.SetSniff(sniff()))
	if err != nil {
		panic(err)
	} else {
		log.Printf("ES Client established")
	}
	clients["bulk"] = client

	cat := elastic.NewCatIndicesService(client)
	catIndices["catIndices"] = cat

	proc, err := clients["bulk"].BulkProcessor().
		Name("worker_1").
		Workers(1).
		FlushInterval(10 * time.Second).
		After(afterBulkCommit).
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	bulkProc["bulk"] = proc
}

func afterBulkCommit(executionId int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
	if response.Errors {
		log.Printf("ERROR: executionId: %d response has errors\n", executionId)
		logErrorDetails(executionId, response.Took, response.Items, requests)
	}
	if err != nil {
		log.Printf("ERROR: executionId: %d encountered an error\n", executionId)
		logErrorDetails(executionId, response.Took, response.Items, requests)
		log.Fatal(err)
	}
}

func logErrorDetails(executionId int64, took int, items []map[string]*elastic.BulkResponseItem, requests []elastic.BulkableRequest) {
	log.Printf("ERROR: executionId: %d, time: %d ms", executionId, took)
	for _, item := range items {
		for _, itemResponse := range item {
			log.Printf("ERROR: executionId: %d, itemResponse: %+v\n", executionId, itemResponse)
			log.Printf("ERROR: executionId: %d, itemResponse.Error: %+v\n", executionId, itemResponse.Error)

		}
	}
	for _, request := range requests {
		log.Printf("executionId: %d, request: %+v\n", executionId, request)
	}
}

func sendToBulker(message []byte) {
	request := elastic.NewBulkIndexRequest().
		Index(indexName()).
		Type("pglog").
		Doc(string(message))
	bulkProc["bulk"].Add(request)
}

func saveToElastic(message []byte) {
	toEs, err := clients["bulk"].Index().
		Index(indexName()).
		Type("pglog").
		BodyString(string(message)).
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", toEs)
}

func indexName() string {
	currentDate := time.Now().Local()
	var buffer bytes.Buffer
	buffer.WriteString("pg-")
	buffer.WriteString(currentDate.Format("2006-01-02"))
	return buffer.String()
}

func bulkStats() elastic.BulkProcessorStats {
	stats := bulkProc["bulk"].Stats()
	fmt.Printf("BulkProcessorStats: %+v\n", stats)
	for i, w := range stats.Workers {
		fmt.Printf("\tBulkProcessorWorkerStats[%d]: %+v\n", i, w)
	}

	return stats
}

func indices() []string {
	indices := make([]string, 0)
	response, err := catIndices["catIndices"].Do(context.Background())

	if err != nil {
		log.Printf("ERROR: could not list indices")
		return indices
	}

	for _, index := range response {
		indices = append(indices, index.Index)
	}
	return indices
}

func elasticURL() string {
	if value, ok := os.LookupEnv("PLS_ELASTIC_URL"); ok {
		return value
	}
	return "http://127.0.0.1:9200"
}

func sniff() bool {
	if env, ok := os.LookupEnv("PLATFORM_ENV"); ok {
		if env == "test" {
			return false
		}
	}

	if value, ok := os.LookupEnv("PLS_ELASTIC_SNIFF"); ok {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return true
}
