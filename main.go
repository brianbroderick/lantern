package main

import (
	"flag"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	elastic "github.com/olivere/elastic/v7"

	"github.com/brianbroderick/logit"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

type batch struct {
	minute time.Time
	sha    string
}

var (
	batchMap        = make(map[batch]*query)
	batchDetailsMap = make(map[batch]*queryDetailStats)
	clients         = make(map[string]*elastic.Client)
	bulkProc        = make(map[string]*elastic.BulkProcessor)
	catIndices      = make(map[string]*elastic.CatIndicesService)
	mutex           = &sync.Mutex{}
	detailsMutex    = &sync.Mutex{}
	redisQueues     = make([]string, 0)
	statsQueues     = make([]string, 0)
	detailArgs      = &queryDetails{}

	// flags
	queuePtr          string
	statsPtr          string
	redisPtr          string
	redisPwPtr        string
	elasticPtr        string
	detailFragmentPtr string
	detailColumnsPtr  string

	// colors for terminal
	blue    = color.New(color.FgBlue).SprintFunc()
	cyan    = color.New(color.FgCyan).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	white   = color.New(color.FgWhite).SprintFunc()

	suppressedCommandTag = map[string]bool{
		"BIND":   true,
		"PARSE":  true,
		"BEGIN":  true,
		"COMMIT": true,
	}
)

func main() {
	flag.StringVar(&queuePtr, "queues", "", "comma separated list of queues that overrides env vars. Can also be set via PLS_REDIS_QUEUES env var")
	flag.StringVar(&statsPtr, "statsQueues", "", "comma separated list of queues for statistics that overrides env vars. Can also be set via PLS_REDIS_STATS_QUEUES env var")
	flag.StringVar(&redisPtr, "redisUrl", "", "Redis URL. Can also set via PLS_REDIS_URL env var")
	flag.StringVar(&redisPwPtr, "redisPassword", "", "Redis password (optional). Can also set via PLS_REDIS_PASSWORD env var")
	flag.StringVar(&elasticPtr, "elasticUrl", "", "Elasticsearch URL. Can also set via PLS_ELASTIC_URL env var")
	flag.StringVar(&detailFragmentPtr, "detailFragment", "", "SQL fragment to look for when parsing out query details")
	flag.StringVar(&detailColumnsPtr, "detailColumn", "", "comma separated list of columns to extract from details and store")

	detailArgs = newQueryDetails(detailFragmentPtr, detailColumnsPtr)

	flag.Parse()
	initialSetup()
	SetupElastic()

	defer bulkProc["bulk"].Close()
	defer clients["bulk"].Stop()

	// Flush to bulkProc every 60 seconds
	ticker := time.NewTicker(time.Second * 60)
	go func() {
		for range ticker.C {
			iterOverQueries()
		}
	}()

	for _, queue := range redisQueues {
		go startRedisBatch(queue, "query")
		time.Sleep(42 * time.Millisecond) // stagger threads hitting Redis
	}
	for _, queue := range statsQueues {
		go startRedisBatch(queue, "stats")
		time.Sleep(42 * time.Millisecond) // stagger threads hitting Redis
	}

	forever := make(chan bool)
	<-forever
}

func initialSetup() {
	setupEnv()
	populateRedisQueues(os.Getenv("PLS_REDIS_QUEUES"))
	populateStatsQueues(os.Getenv("PLS_REDIS_STATS_QUEUES"))
	SetupRedis()
}

func setupEnv() {
	if os.Getenv("PLATFORM_ENV") == "" {
		os.Setenv("PLATFORM_ENV", "prod")
	}

	platformEnv := os.Getenv("PLATFORM_ENV")
	if platformEnv != "prod" && platformEnv != "stage" {
		filename := ".env_" + platformEnv
		err := godotenv.Load(filename)
		if err != nil {
			// logit.Error("%s file not found", filename)
		}
	}
	err := godotenv.Load(".env")
	if err != nil {
		// logit.Error(".env file not found")
	}
}

func populateRedisQueues(queues string) {
	// Override with a flag, if exists
	if queuePtr != "" {
		queues = queuePtr
	}
	if queues == "" {
		redisQueues = append(redisQueues, "postgres")
	} else {
		r := regexp.MustCompile(" ")
		queues = r.ReplaceAllString(queues, "")
		redisQueues = strings.Split(queues, ",")
	}
	uniqueRedisQueues()
	if os.Getenv("PLATFORM_ENV") != "test" {
		logit.Info("Redis Queues: %v", redisQueues)
	}
}

func uniqueRedisQueues() {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range redisQueues {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	redisQueues = list
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

func roundToMinute(minute time.Time) time.Time {
	return minute.Round(time.Minute)
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
