package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	logit "github.com/brianbroderick/logit"
	elastic "gopkg.in/olivere/elastic.v5"
)

// SetupElastic sets up elastic conn
func SetupElastic() {
	// Coming from Docker, sleep a few seconds to make sure ES is running
	if elasticURL() == "http://elasticsearch:9200" {
		logit.Info("Using docker, waiting for ES to spin up")
		time.Sleep(10 * time.Second)
	}
	client, err := elastic.NewClient(elastic.SetURL(elasticURL()), elastic.SetSniff(sniff()))
	if err != nil {
		panic(err)
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

func putTemplate(client *elastic.Client) {
	dat, err := ioutil.ReadFile("./elasticsearch_template.json")
	if err != nil {
		panic(err)
	}

	client.IndexDeleteTemplate("pglog").Do(context.Background())
	_, err = client.IndexPutTemplate("pglog").BodyString(string(dat)).Do(context.Background()) //.Body(dat).Do(context.Background())
	if err != nil {
		panic(err)
	}
}

func afterBulkCommit(executionId int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
	if response.Errors {
		logit.Error("executionId: %d response has errors\n", executionId)
		logErrorDetails(executionId, response.Took, response.Items, requests)
	}
	if err != nil {
		logit.Error("executionId: %d encountered an error\n", executionId)
		logErrorDetails(executionId, response.Took, response.Items, requests)
		logit.Fatal("%v", err)
	}
}

func logErrorDetails(executionId int64, took int, items []map[string]*elastic.BulkResponseItem, requests []elastic.BulkableRequest) {

	for _, item := range items {
		for _, itemResponse := range item {
			if itemResponse.Error != nil {
				logit.Error("logErrorDetails:2 executionId: %d, itemResponse: %+v\n", executionId, itemResponse)
				logit.Error("logErrorDetails:3 executionId: %d, itemResponse.Error: %+v\n", executionId, itemResponse.Error)
			}
		}
	}
	for _, request := range requests {
		if executionId == 1 {
			logit.Info("logErrorDetails:4 executionId: %d, request: %+v\n---\n", executionId, request)
		}
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
		logit.Error("could not list indices")
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
