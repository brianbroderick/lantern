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

const longForm = "2006-12-06T23:13:33.242+0000"

type query struct {
	uniqueSha     string // sha of uniqueStr and preparedStep (if available)
	uniqueStr     string // usually the normalized query
	commandTag    string
	errorSeverity string
	logType       string
	notes         string
	message       string
	grokQuery     string
	prepared      string
	preparedStep  string
	query         string
	tempTable     int64
	timestamp     time.Time
	totalCount    int32
	totalDuration float64
	vacuumTable   string
	data          map[string]*json.RawMessage
}

func newQuery(b []byte) (*query, error) {
	var q = new(query)
	q.totalCount = 1

	if err := json.Unmarshal(b, &q.data); err != nil {
		return nil, err
	}

	if source, pres := q.data["command_tag"]; pres {
		if err := json.Unmarshal(*source, &q.commandTag); err != nil {
			return nil, err
		}
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

	var tempTime string
	if source, pres := q.data["@timestamp"]; pres {
		if err := json.Unmarshal(*source, &tempTime); err != nil {
			return nil, err
		}
	}
	q.timestamp, _ = time.Parse(longForm, tempTime)

	// If it's an error, use the error code as the uniqueStr
	if q.errorSeverity == "ERROR" || q.errorSeverity == "FATAL" {
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
	// DELETE message once we've debugged all the messages.
	// delete(q.data, "message")

	return q, nil
}

func regexMessage(message string) map[string]string {
	// Query regexp
	r := regexp.MustCompile(`(?s)duration: (?P<duration>\d+\.\d+) ms\s+(?P<preparedStep>\w+)\s*?(?P<prepared>.*?)?:\s*(?P<grokQuery>.*)`)
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

	//temporary file: path "base/pgsql_tmp/pgsql_tmp14938.66", size 708064
	r = regexp.MustCompile(`(?s)temporary file: path ".*?", size\s+(?P<tempTable>\d+).*`)
	match = r.FindStringSubmatch(message)
	result = make(map[string]string)

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
		result["logType"] = "connection_received"
		return result
	}

	// disconnection: session time: 0:00:00.074 user=q55cd17435 database= host=10.0.1.168 port=56544
	r = regexp.MustCompile(`(?s)disconnection:.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "disconnection"
		return result
	}

	// replication connection authorized: user=q55cd17435 SSL enabled (protocol=TLSv1.2, cipher=ECDHE-RSA-AES256-GCM-SHA384, compression=off)
	r = regexp.MustCompile(`(?s)replication connection authorized:.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "connection_replication"
		return result
	}

	// checkpoint starting: time
	r = regexp.MustCompile(`(?s)checkpoint (?P<actionCheckpoint>starting|complete):.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		for i, name := range r.SubexpNames() {
			if i != 0 {
				result[name] = match[i]
			}
		}
		result["logType"] = "checkpoint_" + result["actionCheckpoint"]
		return result
	}

	//automatic vacuum of table "app.public.api_clients":.*
	// or automatic analyze of table "app.public.api_clients" system usage: CPU 0.00s/0.02u sec elapsed 0.15 sec
	r = regexp.MustCompile(`(?s)automatic (?P<action>vacuum|analyze) of table "(?P<table>.*?)":.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		for i, name := range r.SubexpNames() {
			if i != 0 {
				result[name] = match[i]
			}
		}
		result["notes"] = message
		result["logType"] = result["action"] + "_table " + result["table"]
		return result
	}

	//could not receive data from client: Connection reset by peer
	r = regexp.MustCompile(`(?s)could not receive data.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "connection_reset"
		return result
	}

	result["unknownMessage"] = message
	return result
}

func parseMessage(q *query) error {
	result := make(map[string]string)

	if q.commandTag == "UPDATE waiting" || q.commandTag == "INSERT waiting" {
		grokQuery, err := unmarshalQuery(q)
		if err != nil {
			return err
		}

		result["grokQuery"] = grokQuery
		q.notes = q.message
	} else if q.commandTag == "authentication" {
		result["logType"] = "connection_authorized"
	} else {
		result = regexMessage(q.message)
	}

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

		// When there's a temp table, the "query" field is passed
		if result["tempTable"] != "" {
			grokQuery, err := unmarshalQuery(q)
			if err != nil {
				return err
			}

			q.tempTable, err = strconv.ParseInt(result["tempTable"], 10, 64)
			if err != nil {
				return err
			}
			result["grokQuery"] = grokQuery
		}

		if result["grokQuery"] != "" {
			q.preparedStep = result["preparedStep"]
			q.prepared = strings.TrimSpace(result["prepared"])
			q.query = result["grokQuery"]

			pgQuery, err := normalizeQuery(result["grokQuery"])
			if err != nil {
				return err
			}

			q.uniqueStr = string(pgQuery)
		}

		if result["logType"] != "" {
			q.notes = q.message
			q.uniqueStr = result["logType"]
		}

		if q.commandTag == "DEALLOCATE" {
			q.uniqueStr = "deallocate"
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

	// uniqueSha
	if q.uniqueSha != "" {
		err = marshalString(q, q.uniqueSha, "unique_sha")
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

	// tempTable
	if q.tempTable > 0 {
		b, err := json.Marshal(q.tempTable)
		if err != nil {
			return nil, err
		}
		tempTable := json.RawMessage(b)
		q.data["temp_table_size"] = &tempTable
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
	mutex.Lock()
	_, ok := batchMap[batch{roundMin, q.uniqueSha}]
	if ok == true {
		batchMap[batch{roundMin, q.uniqueSha}].totalCount++
		batchMap[batch{roundMin, q.uniqueSha}].totalDuration += q.totalDuration
	} else {
		batchMap[batch{roundMin, q.uniqueSha}] = q
	}
	mutex.Unlock()
}

func iterOverQueries() {
	var duration time.Duration
	now := currentMinute()
	mutex.Lock()
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
	mutex.Unlock()
}

func unmarshalQuery(q *query) (string, error) {
	var grokQuery string
	// var err error
	if source, pres := q.data["query"]; pres {
		if err := json.Unmarshal(*source, &grokQuery); err != nil {
			return "", err
		}
	}
	return grokQuery, nil
}
