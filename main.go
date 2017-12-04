package main

import (
	"log"
	"math"
	"os"
	"sync"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/joho/godotenv"
)

type batch struct {
	minute time.Time
	sha    string
}

var (
	batchMap = make(map[batch]*query)
	clients  = make(map[string]*elastic.Client)
	bulkProc = make(map[string]*elastic.BulkProcessor)
	mutex    = &sync.Mutex{}
)

func main() {
	initialSetup()
	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()

	// Flush to bulkProc every 30 seconds
	ticker := time.NewTicker(time.Second * 30)
	go func() {
		for _ = range ticker.C {
			iterOverQueries()
		}
	}()

	go startRedisBatch()

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
		err := godotenv.Load(filename)
		if err != nil {
			log.Printf("INFO: %s file not found", filename)
		}
	}
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

func lastMinute() time.Time {
	return currentMinute().Add(-1 * time.Minute)
}

func round(val float64, roundOn float64, places int) float64 {

	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)

	var round float64
	if val > 0 {
		if div >= roundOn {
			round = math.Ceil(digit)
		} else {
			round = math.Floor(digit)
		}
	} else {
		if div >= roundOn {
			round = math.Floor(digit)
		} else {
			round = math.Ceil(digit)
		}
	}

	return round / pow
}
