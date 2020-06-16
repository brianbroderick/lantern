package main

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"time"

	logit "github.com/brianbroderick/logit"
)

// This doesn't come from PG's Redislog. It's meant to come from a non-PG source such as an application
// mimicking the PG RedisLog payload. For such an application, it can also send a stats payload to
// show how many messages failed to send or were suppressed. See sample_payloads/stats.json for an example

type stats struct {
	applicationName string
	interval        int64
	intervalUnit    string
	libraryVersion  string
	timestamp       time.Time
	failedEvents    int64
	ignoredEvents   int64
	redisKey        string
	data            map[string]*json.RawMessage
}

func newStats(b []byte, redisKey string) (*stats, error) {
	var q = new(stats)

	if err := json.Unmarshal(b, &q.data); err != nil {
		return nil, err
	}

	str, err := json.Marshal(redisKey)
	if err != nil {
		return nil, err
	}

	rawMarshal := json.RawMessage(str)
	q.data["redis_key"] = &rawMarshal

	return q, nil
}

func populateStatsQueues(queues string) {
	// Override with a flag, if exists
	if queuePtr != "" {
		queues = queuePtr
	}

	if os.Getenv("PLATFORM_ENV") == "test" {
		queues = "stats"
	}

	if queues != "" {
		r := regexp.MustCompile(" ")
		queues = r.ReplaceAllString(queues, "")
		statsQueues = strings.Split(queues, ",")

		uniqueStatsQueues()
		if os.Getenv("PLATFORM_ENV") != "test" {
			logit.Info("Stats Queues: %v", statsQueues)
		}
	}
}

func uniqueStatsQueues() {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range statsQueues {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	statsQueues = list
}
