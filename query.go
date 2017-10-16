package main

import "encoding/json"

type query struct {
	query           string
	normalizedQuery string
	duration        float64
	message         string
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

	var q = new(query)
	q.data = data
	q.message = message

	return q, nil
}
