package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"

	logit "github.com/brettallred/go-logit"
)

type query struct {
	uniqueSha            string
	query                string
	normalizedQuery      string
	duration             float64
	count                int32
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
		batchMap[batch{roundMin, q.uniqueSha}].count++
		batchMap[batch{roundMin, q.uniqueSha}].duration += q.duration
	} else {
		batchMap[batch{roundMin, q.uniqueSha}] = q
	}
}

func iterOverQueries() {
	var duration time.Duration
	now := currentMinute()
	for k := range batchMap {
		duration = now.Sub(k.minute)
		if duration >= (1 * time.Minute) {
			batchMap[k].marshalAgg()
			data, err := json.Marshal(batchMap[k].data)
			if err != nil {
				logit.Error("Error marshalling data: %e", err.Error())
			}
			sendToBulker(data)
			delete(batchMap, k)
		} else {
			fmt.Printf("\n els[%s]\n", duration)
		}

		// fmt.Printf("\nkey[%s]\n", k)
		// fmt.Printf("\nkey[%s] value[%v+]\n", k, batchMap[k])
	}
}

func parseMessage(q *query) error {
	r := regexp.MustCompile(`duration: (?P<duration>\d+\.\d+) ms\s+(?P<preparedStep>[a-zA-Z0-9]+)\s+(?P<prepared>.*):\s*(?P<query>.*)`)
	match := r.FindStringSubmatch(q.message)
	result := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 {
			result[name] = match[i]
		}
	}

	duration, err := strconv.ParseFloat(result["duration"], 64)
	if err != nil {
		return err
	}
	q.duration = duration
	q.count = 1
	q.preparedStep = result["preparedStep"]
	q.prepared = result["prepared"]
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
	b, err := json.Marshal(q.count)
	if err != nil {
		return nil, err
	}
	rawCount := json.RawMessage(b)
	q.data["count"] = &rawCount

	// duration
	b, err = json.Marshal(q.duration)
	if err != nil {
		return nil, err
	}
	rawDuration := json.RawMessage(b)
	q.data["duration"] = &rawDuration

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
	err = marshalString(q, q.uniqueSha, "normalized_sha")
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
