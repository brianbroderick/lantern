package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	logit "github.com/brettallred/go-logit"
)

type query struct {
	uniqueSha            string
	query                string
	normalizedQuery      string
	totalDuration        float64
	totalCount           int32
	message              string
	preparedStep         string
	prepared             string
	virtualTransactionID string
	data                 map[string]*json.RawMessage
}

func newQuery(b []byte) (*query, error) {
	var data map[string]*json.RawMessage
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	var message string
	if source, pres := data["message"]; pres {
		if err := json.Unmarshal(*source, &message); err != nil {
			return nil, err
		}
	}

	var que = new(query)
	que.data = data
	que.message = message
	if err := parseMessage(que); err != nil {
		return nil, err
	}

	return que, nil
}

func addToQueries(roundMin time.Time, q *query) {
	_, ok := batchMap[batch{roundMin, q.uniqueSha}]
	if ok == true {
		batchMap[batch{roundMin, q.uniqueSha}].totalCount++
		batchMap[batch{roundMin, q.uniqueSha}].totalDuration += q.totalDuration
	} else {
		batchMap[batch{roundMin, q.uniqueSha}] = q
	}
}

func iterOverQueries() {
	logit.Info(" Checking if queries are ready for ES Bulk Processor")
	var duration time.Duration
	now := currentMinute()
	for k := range batchMap {
		duration = now.Sub(k.minute)
		if duration >= (1 * time.Minute) {
			logit.Info(" Sending queries to ES Bulk Processor")
			batchMap[k].marshalAgg()
			data, err := json.Marshal(batchMap[k].data)
			if err != nil {
				logit.Error(" Error marshalling data: %e", err.Error())
			}
			sendToBulker(data)
			delete(batchMap, k)
		}
	}
}

func regexMessage(message string) (map[string]string, error) {
	r := regexp.MustCompile(`(?s)duration: (?P<duration>\d+\.\d+) ms\s+(?P<preparedStep>\w+)\s*?(?P<prepared>.*?)?:\s*(?P<query>.*)`)
	match := r.FindStringSubmatch(message)
	result := make(map[string]string)

	if len(match) > 0 {
		for i, name := range r.SubexpNames() {
			if i != 0 {
				result[name] = match[i]
			}
		}
		return result, nil
	}
	return nil, errors.New("regexMessage: match not found: " + message)

}

func parseMessage(q *query) error {
	// logit.Info(" Message: %s\n\n", q.message)

	result, err := regexMessage(q.message)
	if err != nil {
		return err
	}

	duration, err := strconv.ParseFloat(result["duration"], 64)
	if err != nil {
		return err
	}
	q.totalDuration = duration
	q.totalCount = 1
	q.preparedStep = result["preparedStep"]
	q.prepared = strings.TrimSpace(result["prepared"])
	q.query = result["query"]

	pgQuery, err := normalizeQuery(result["query"])
	if err != nil {
		return err
	}

	q.normalizedQuery = string(pgQuery)
	q.shaUnique()
	q.marshal()

	return nil
}

// creates a sha of the prepared step and normalized query
func (q *query) shaUnique() {
	h := sha1.New()
	io.WriteString(h, q.normalizedQuery)
	io.WriteString(h, q.preparedStep)
	q.uniqueSha = hex.EncodeToString(h.Sum(nil))
}

func (q *query) marshalAgg() ([]byte, error) {
	// count
	b, err := json.Marshal(q.totalCount)
	if err != nil {
		return nil, err
	}
	rawCount := json.RawMessage(b)
	q.data["total_count"] = &rawCount

	// duration
	b, err = json.Marshal(q.totalDuration)
	if err != nil {
		return nil, err
	}
	rawDuration := json.RawMessage(b)
	q.data["total_duration_ms"] = &rawDuration

	return json.Marshal(q.data)
}

func (q *query) marshal() ([]byte, error) {
	// preparedStep
	err := marshalString(q, q.preparedStep, "prepared_step")
	if err != nil {
		return nil, err
	}

	// prepared
	err = marshalString(q, q.prepared, "prepared")
	if err != nil {
		return nil, err
	}

	// query
	err = marshalString(q, q.query, "query")
	if err != nil {
		return nil, err
	}

	// normalizedQuery
	err = marshalString(q, q.normalizedQuery, "normalized_query")
	if err != nil {
		return nil, err
	}

	// uniqueSha
	err = marshalString(q, q.uniqueSha, "normalized_prepared_step_sha")
	if err != nil {
		return nil, err
	}

	delete(q.data, "message")

	return json.Marshal(q.data)
}

func marshalString(q *query, strToMarshal string, dataKey string) error {
	b, err := json.Marshal(strToMarshal)
	if err != nil {
		return err
	}
	rawMarshal := json.RawMessage(b)
	q.data[dataKey] = &rawMarshal

	return nil
}
