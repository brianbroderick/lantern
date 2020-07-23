package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	logit "github.com/brianbroderick/logit"
	elastic "github.com/olivere/elastic/v7"
)

// SetupElastic sets up elastic conn
func SetupElastic() {
	if os.Getenv("PLATFORM_ENV") != "test" {
		logit.Info("Elastic URL: %s\n", elasticURL())
	}

	// Coming from Docker, sleep a few seconds to make sure ES is running
	if elasticURL() == "http://elasticsearch:9200" {
		logit.Info("Using docker, waiting for ES to spin up")
		time.Sleep(10 * time.Second)
	}
	client := elasticClientFactory()

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

func elasticClientFactory() *elastic.Client {
	if isUsingBasicAuth() {
		logit.Info("Elastic client is using basic auth.\n")
		// ignore self-signed certs locally
		if !validateCertificates() {
			logit.Info("Elastic client is using not validating certificates.\n")
			httpClient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}
			client, err := elastic.NewClient(
				elastic.SetHttpClient(httpClient),
				elastic.SetBasicAuth(elasticUser(), elasticPassword()),
				elastic.SetURL(elasticURL()),
				elastic.SetSniff(sniff()),
			)
			if err != nil {
				panic(err)
			}
			return client
		} else {
			client, err := elastic.NewClient(
				elastic.SetBasicAuth(elasticUser(), elasticPassword()),
				elastic.SetURL(elasticURL()),
				elastic.SetSniff(sniff()),
			)
			if err != nil {
				panic(err)
			}
			return client
		}
	} else {
		logit.Info("Elastic client has no auth\n")
		client, err := elastic.NewClient(
			elastic.SetURL(elasticURL()),
			elastic.SetSniff(sniff()),
		)
		if err != nil {
			panic(err)
		}
		return client
	}
}

func putTemplate(client *elastic.Client) {
	dat, err := ioutil.ReadFile("./elasticsearch_template.json")
	if err != nil {
		panic(err)
	}

	exists, err := client.IndexTemplateExists("pglog").Do(context.Background())
	if err != nil {
		panic(err)
	}

	if exists {
		client.IndexDeleteTemplate("pglog").Do(context.Background())
	}

	res, err := client.IndexPutTemplate("pglog").BodyString(string(dat)).Do(context.Background()) //.Body(dat).Do(context.Background())
	if err != nil {
		panic(err)
	}

	logit.Info("Created 'pglog' template template: %t", res.Acknowledged)
}

func afterBulkCommit(executionId int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
	logit.Info(yellow("Posted %d records to ES"), len(requests))

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
				logit.Error("logErrorDetails executionId: %d, itemResponse: %+v\n", executionId, itemResponse)
				logit.Error("logErrorDetails executionId: %d, itemResponse.Error: %+v\n", executionId, itemResponse.Error)
			}
		}
	}
	// This prints all the requests in the batch. I haven't figured out how to tie this back to the actual errors.
	// This isn't super helpful when most of the queries work. Uncomment out if needed.
	// for _, request := range requests {
	// 	logit.Info("logErrorDetails executionId: %d, request: %+v\n---\n", executionId, request)
	// }
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
	// If flag is set, use that.
	if elasticPtr != "" {
		return elasticPtr
	}
	// Next use an environment variable, if it's set
	if value, ok := os.LookupEnv("PLS_ELASTIC_URL"); ok {
		return value
	}
	// Lastly return default
	return "https://localhost:9200"
}


func validateCertificates() bool  {
	if value, ok := os.LookupEnv("PLS_VALIDATE_CERTIFICATES"); ok {
		return strings.EqualFold(value, "true")
	}
	return true
}

func isUsingBasicAuth() bool  {
	if value, ok := os.LookupEnv("PLS_ELASTIC_BASIC_AUTH"); ok {
		return strings.EqualFold(value, "true")
	}
	return len(elasticPassword()) > 0
}

func elasticUser() string {
	if value, ok := os.LookupEnv("PLS_ELASTIC_USERNAME"); ok {
		return value
	}
	return "elastic"
}

func elasticPassword() string {
	if value, ok := os.LookupEnv("PLS_ELASTIC_PASSWORD"); ok {
		return value
	}
	return ""
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
