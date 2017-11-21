package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	logit "github.com/brettallred/go-logit"
)

type query struct {
	uniqueSha     string // sha of uniqueStr and preparedStep (if available)
	uniqueStr     string // usually the normalized query
	notes         string
	errorSeverity string
	message       string
	totalDuration float64
	totalCount    int32
	query         string
	preparedStep  string
	prepared      string
	logType       string
	data          map[string]*json.RawMessage
}

func newQuery(b []byte) (*query, error) {
	var q = new(query)
	q.totalCount = 1

	if err := json.Unmarshal(b, &q.data); err != nil {
		return nil, err
	}

	if source, pres := q.data["error_severity"]; pres {
		if err := json.Unmarshal(*source, &q.errorSeverity); err != nil {
			return nil, err
		}
	}

	if source, pres := q.data["message"]; pres {
		if err := json.Unmarshal(*source, &q.message); err != nil {
			return nil, err
		}
	}

	// If it's an error, use the error code as the uniqueStr
	if q.errorSeverity == "ERROR" {
		if source, pres := q.data["sql_state_code"]; pres {
			if err := json.Unmarshal(*source, &q.uniqueStr); err != nil {
				return nil, err
			}
		}
		q.notes = q.message
	} else { // assumed the errorSeverity is "LOG"
		if err := parseMessage(q); err != nil {
			return nil, err
		}
	}

	q.shaUnique()
	q.marshal()
	delete(q.data, "message")

	return q, nil
}

func regexMessage(message string) map[string]string {
	// Query regexp
	r := regexp.MustCompile(`(?s)duration: (?P<duration>\d+\.\d+) ms\s+(?P<preparedStep>\w+)\s*?(?P<prepared>.*?)?:\s*(?P<query>.*)`)
	match := r.FindStringSubmatch(message)
	result := make(map[string]string)

	if len(match) > 0 {
		for i, name := range r.SubexpNames() {
			if i != 0 {
				result[name] = match[i]
			}
		}
		return result
	}

	// connection received: host=10.0.1.168 port=38634
	r = regexp.MustCompile(`(?s)connection received:.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "connection"
		return result
	}

	// replication connection authorized: user=q55cd17435 SSL enabled (protocol=TLSv1.2, cipher=ECDHE-RSA-AES256-GCM-SHA384, compression=off)
	r = regexp.MustCompile(`(?s)replication connection authorized:.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "replication_connection"
		return result
	}

	// disconnection: session time: 0:00:00.074 user=q55cd17435 database= host=10.0.1.168 port=56544
	r = regexp.MustCompile(`(?s)disconnection:.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "disconnection"
		return result
	}

	// checkpoint starting: time
	r = regexp.MustCompile(`(?s)checkpoint starting:.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "checkpoint"
		return result
	}

	result["unknownMessage"] = message
	return result
}

func parseMessage(q *query) error {
	result := regexMessage(q.message)

	if result["unknownMessage"] != "" {
		// unknownMessage
		q.uniqueStr = result["unknownMessage"]
		err := marshalString(q, result["unknownMessage"], "unknown_message")
		if err != nil {
			return err
		}
	} else {
		if result["duration"] != "" {
			duration, err := strconv.ParseFloat(result["duration"], 64)
			if err != nil {
				return err
			}
			q.totalDuration = duration
		}

		if result["query"] != "" {
			q.preparedStep = result["preparedStep"]
			q.prepared = strings.TrimSpace(result["prepared"])
			q.query = result["query"]

			pgQuery, err := normalizeQuery(result["query"])
			if err != nil {
				return err
			}

			q.uniqueStr = string(pgQuery)
		}

		if result["logType"] != "" {
			q.notes = q.message
			q.uniqueStr = result["logType"]
		}
	}

	return nil
}

// creates a sha of the prepared step and normalized query
func (q *query) shaUnique() {
	h := sha1.New()
	io.WriteString(h, q.uniqueStr)
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

	// duration rounded to 5 decimal points
	b, err = json.Marshal(round(q.totalDuration, 0.5, 5))
	if err != nil {
		return nil, err
	}
	rawDuration := json.RawMessage(b)
	q.data["total_duration_ms"] = &rawDuration

	return json.Marshal(q.data)
}

func (q *query) marshal() ([]byte, error) {
	var err error

	// preparedStep
	if q.preparedStep != "" {
		err = marshalString(q, q.preparedStep, "prepared_step")
		if err != nil {
			return nil, err
		}
	}

	// prepared
	if q.prepared != "" {
		err = marshalString(q, q.prepared, "prepared")
		if err != nil {
			return nil, err
		}
	}

	// query
	if q.query != "" {
		err = marshalString(q, q.query, "query")
		if err != nil {
			return nil, err
		}
	}

	// uniqueStr
	err = marshalString(q, q.uniqueStr, "unique_string")
	if err != nil {
		return nil, err
	}

	// errorMessage
	if q.notes != "" {
		err = marshalString(q, q.notes, "notes")
		if err != nil {
			return nil, err
		}
	}

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
	var duration time.Duration
	now := currentMinute()
	for k := range batchMap {
		duration = now.Sub(k.minute)
		if duration >= (1 * time.Minute) {
			logit.Info(" Sending %s to ES Bulk Processor", k.sha)
			if k.sha == "" {
				logit.Info("%s", batchMap[k].data)
			}
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
