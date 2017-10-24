package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"regexp"
	"strconv"
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

func addToQueries(q *query) {
	_, ok := queryMap[q.uniqueSha]
	if ok == true {
		queryMap[q.uniqueSha].count++
		queryMap[q.uniqueSha].duration += q.duration
	} else {
		queryMap[q.uniqueSha] = q
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

func (q *query) marshal() ([]byte, error) {
	// duration
	b, err := json.Marshal(q.duration)
	if err != nil {
		return nil, err
	}
	rawDuration := json.RawMessage(b)
	q.data["duration"] = &rawDuration

	// preparedStep
	b, err = json.Marshal(q.preparedStep)
	if err != nil {
		return nil, err
	}
	rawPreparedStep := json.RawMessage(b)
	q.data["prepared_step"] = &rawPreparedStep

	// prepared
	b, err = json.Marshal(q.prepared)
	if err != nil {
		return nil, err
	}
	rawPrepared := json.RawMessage(b)
	q.data["prepared"] = &rawPrepared

	// query
	b, err = json.Marshal(q.query)
	if err != nil {
		return nil, err
	}
	rawQuery := json.RawMessage(b)
	q.data["query"] = &rawQuery

	// normalizedQuery
	b, err = json.Marshal(q.normalizedQuery)
	if err != nil {
		return nil, err
	}
	rawNormalizedQuery := json.RawMessage(b)
	q.data["normalized_query"] = &rawNormalizedQuery

	// uniqueSha
	b, err = json.Marshal(q.uniqueSha)
	if err != nil {
		return nil, err
	}
	rawUniqueSha := json.RawMessage(b)
	q.data["normalized_sha"] = &rawUniqueSha

	delete(q.data, "message")

	return json.Marshal(q.data)
}
