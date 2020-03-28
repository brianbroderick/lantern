package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

// SetupElastic sets up elastic conn
func SetupElastic() {
	client, err := elastic.NewClient(elastic.SetURL(elasticURL()), elastic.SetSniff(sniff()))
	if err != nil {
		panic(err)
	}
	clients["bulk"] = client

	proc, err := clients["bulk"].BulkProcessor().
		Name("worker_1").
		Workers(1).
		FlushInterval(10 * time.Second).
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	bulkProc["bulk"] = proc
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
	return bulkProc["bulk"].Stats()
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
