package main

import (
	"context"
	"fmt"

	elastic "gopkg.in/olivere/elastic.v5"
)

func saveToElastic() {
	client, err := elastic.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Stop()

	exists, err := client.IndexExists("twitter").Do(context.Background())
	if err != nil {
		fmt.Println("no bueno")
	}
	if !exists {
		fmt.Println("doesn't exist")
	}

}
