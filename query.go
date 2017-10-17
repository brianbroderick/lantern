package main

import (
	"encoding/json"
	"regexp"
	"strconv"
)

type query struct {
	query           string
	normalizedQuery string
	duration        float64
	message         string
	commandTag      string
	prepared        string
	data            map[string]*json.RawMessage
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

func parseMessage(q *query) error {
	r := regexp.MustCompile(`duration: (?P<duration>\d+\.\d+) ms\s+(?P<commandTag>[a-zA-Z0-9]+)\s+(?P<prepared>.*):\s*(?P<query>.*)`)
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
	q.commandTag = result["commandTag"]
	q.prepared = result["prepared"]
	q.query = result["query"]
	pgQuery, err := normalizeQuery(result["query"])
	if err != nil {
		return err
	}
	q.normalizedQuery = string(pgQuery)

	return nil
}
