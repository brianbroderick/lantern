package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

func saveToElastic(message []byte) {
	client, err := elastic.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Stop()

	currentDate := time.Now().Local()
	var buffer bytes.Buffer
	buffer.WriteString("pg-")
	buffer.WriteString(currentDate.Format("2006-01-02"))

	toEs, err := client.Index().
		Index(buffer.String()).
		Type("pglog").
		BodyString(string(message)).
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", toEs)
}
