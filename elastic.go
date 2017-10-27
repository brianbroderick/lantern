package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

// SetupElastic sets up elastic conn
func SetupElastic() {
	client, err := elastic.NewClient()
	if err != nil {
		panic(err)
	}
	clients["bulk"] = client

	proc, err := clients["bulk"].BulkProcessor().
		Name("worker_1").
		Workers(1).
		FlushInterval(60 * time.Second).
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	bulkProc["bulk"] = proc
}

func saveToElastic(message []byte) {
	currentDate := time.Now().Local()
	var buffer bytes.Buffer
	buffer.WriteString("pg-")
	buffer.WriteString(currentDate.Format("2006-01-02"))

	toEs, err := clients["bulk"].Index().
		Index(buffer.String()).
		Type("pglog").
		BodyString(string(message)).
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", toEs)
}

// // SetupElastic starts the bulk processor
// func SetupElastic() {
// 	client, err := elastic.NewClient()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer client.Stop()

// 	p, err := client.BulkProcessor().
// 		Name("worker_1").
// 		Workers(1).
// 		FlushInterval(60 * time.Second).
// 		Do(context.Background())
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer p.Close()

// }
