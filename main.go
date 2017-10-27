package main

import (
	"log"
	"os"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/joho/godotenv"
)

type batch struct {
	minute time.Time
	query  string
}

var (
	batchMap = make(map[batch]*query)
	clients  = make(map[string]*elastic.Client)
	bulkProc = make(map[string]*elastic.BulkProcessor)
)

func main() {
	initialSetup()

	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()

	forever := make(chan bool)
	<-forever
}

func initialSetup() {
	setupEnv()
	SetupRedis()
	SetupElastic()
}

func setupEnv() {
	platformEnv := os.Getenv("PLATFORM_ENV")
	if platformEnv != "prod" && platformEnv != "stage" {
		filename := ".env_" + platformEnv
		log.Printf("INFO: loading file %s", filename)
		err := godotenv.Load(filename)
		if err != nil {
			log.Printf("INFO: %s file not found", filename)
		}
	}
	log.Printf("INFO: loading file .env")
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("INFO: .env file not found")
	}
}

func redisKey() string {
	platformEnv := os.Getenv("PLATFORM_ENV")
	if platformEnv != "test" {
		return "postgres"
	}

	return "postgres_test"
}

func currentMinute() time.Time {
	return time.Now().UTC().Round(time.Minute)
}
