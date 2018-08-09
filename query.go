package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	logit "github.com/brettallred/go-logit"
	"github.com/fatih/color"
)

const longForm = "2006-01-02T15:04:05.999+0000"

type query struct {
	uniqueSha       string // sha of uniqueStr and preparedStep (if available)
	uniqueStr       string // usually the normalized query
	codeAction      string
	codeApplication string
	codeController  string
	codeJob         string
	codeLine        string
	codeSource      []map[string]string
	codeSourceMap   map[string]map[string]string
	comments        []string
	commandTag      string
	weekday         string
	weekdayInt      int64
	errorSeverity   string
	logType         string
	notes           string
	minDuration     float64
	maxDuration     float64
	message         string
	grokQuery       string
	prepared        string
	preparedStep    string
	query           string
	redisKey        string
	tempTable       int64
	timestamp       time.Time
	totalCount      int32
	totalDuration   float64
	vacuumTable     string
	data            map[string]*json.RawMessage
}

func newQuery(b []byte, redisKey string) (*query, bool, error) {
	var q = new(query)
	q.codeSourceMap = make(map[string]map[string]string)
	q.totalCount = 1
	q.redisKey = redisKey

	if err := json.Unmarshal(b, &q.data); err != nil {
		return nil, false, err
	}

	if source, pres := q.data["command_tag"]; pres {
		if err := json.Unmarshal(*source, &q.commandTag); err != nil {
			return nil, false, err
		}
	}

	// suppress import under certain conditions like unneccessary command_tags
	if suppressedCommandTag[q.commandTag] {
		return q, true, nil
	}

	if source, pres := q.data["error_severity"]; pres {
		if err := json.Unmarshal(*source, &q.errorSeverity); err != nil {
			return nil, false, err
		}
	}

	if source, pres := q.data["message"]; pres {
		if err := json.Unmarshal(*source, &q.message); err != nil {
			return nil, false, err
		}
	}

	var tempTime string
	if source, pres := q.data["@timestamp"]; pres {
		if err := json.Unmarshal(*source, &tempTime); err != nil {
			return nil, false, err
		}
	}
	q.timestamp, _ = time.Parse(longForm, tempTime)

	// If it's an error, use the error code as the uniqueStr
	if q.errorSeverity == "ERROR" || q.errorSeverity == "FATAL" || q.errorSeverity == "WARNING" {
		if source, pres := q.data["sql_state_code"]; pres {
			if err := json.Unmarshal(*source, &q.uniqueStr); err != nil {
				return nil, false, err
			}
		}
		q.notes = q.message
	} else { // assumed the errorSeverity is "LOG"
		extractComments(q)
		enumComments(q)

		if err := parseMessage(q); err != nil {
			return nil, false, err
		}
	}

	q.shaUnique()
	q.marshal()
	// DELETE message once we've debugged all the messages.
	delete(q.data, "message")

	return q, false, nil
}

func extractComments(q *query) {
	// find comments
	r := regexp.MustCompile(`(/\*.*?:.*?\*/)`)
	match := r.FindAllStringSubmatch(q.message, -1)

	// put comments in their own slice
	if len(match) > 0 {
		comments := make([]string, len(match))
		for i, matches := range match {
			comments[i] = matches[i]
		}
		q.comments = comments

		// remove comments from message
		re := regexp.MustCompile(`(\s*/\*.*?\*/\s*)`)
		q.message = re.ReplaceAllString(q.message, "")
	}
}

func enumComments(q *query) {
	for _, comments := range q.comments {
		q.codeSourceMap = parseComments(q, comments, q.codeSourceMap)
	}
}

// extract code location from comments, return as a map for uniqueness
func parseComments(q *query, comment string, uniqMap map[string]map[string]string) map[string]map[string]string {
	h := sha1.New()
	io.WriteString(h, comment)
	mapSha := hex.EncodeToString(h.Sum(nil))

	r := regexp.MustCompile(`(/\*|\*/)`)
	re := regexp.MustCompile(`(?P<key>.*?):(?P<value>.*)`)
	result := make(map[string]string)

	comment = r.ReplaceAllString(comment, "")

	parts := strings.Split(comment, ",")
	codeSource := make(map[string]string)
	for _, item := range parts {
		match := re.FindStringSubmatch(item)

		if len(match) > 0 {

			for i, name := range re.SubexpNames() {
				if i != 0 {
					result[name] = match[i]
				}
			}

			codeSource[result["key"]] = result["value"]

			switch result["key"] {
			case "application":
				q.codeApplication = result["value"]
			case "controller":
				q.codeController = result["value"]
			case "action":
				q.codeAction = result["value"]
			case "line":
				q.codeLine = result["value"]
			case "job":
				q.codeJob = result["value"]
			}
		}
	}
	uniqMap[mapSha] = codeSource
	return uniqMap
}

func regexMessage(message string) map[string]string {
	result := make(map[string]string)

	// Query regexp
	r := regexp.MustCompile(`(?s)duration: (?P<duration>\d+\.\d+) ms\s+(?P<preparedStep>\w+)\s*?(?P<prepared>.*?)?:\s*(?P<grokQuery>.*)`)
	match := r.FindStringSubmatch(message)

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

	// checkpoint or restartpoint starting or completing
	r = regexp.MustCompile(`(?s)(?P<actionPoint>checkpoint|restartpoint) (?P<actionCheckpoint>starting|complete):.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		for i, name := range r.SubexpNames() {
			if i != 0 {
				result[name] = match[i]
			}
		}
		result["logType"] = result["actionPoint"] + "_" + result["actionCheckpoint"]
		return result
	}

	//automatic vacuum of table "app.public.api_clients":.*
	// or automatic analyze of table "app.public.api_clients" system usage: CPU 0.00s/0.02u sec elapsed 0.15 sec
	r = regexp.MustCompile(`(?s)automatic (?P<action>vacuum|analyze) of table "(?P<table>.*?)".*`)
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

	//recovery restart point at 645/28313308
	r = regexp.MustCompile(`(?s)recovery restart point.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "recovery_restart_point"
		return result
	}

	//could not receive data from client: Connection reset by peer
	r = regexp.MustCompile(`(?s).*Connection reset by peer.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "connection_reset"
		return result
	}

	//unexpected EOF on client connection with an open transaction
	r = regexp.MustCompile(`(?s)unexpected EOF.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "unexpected_eof"
		return result
	}

	//logical decoding found consistent point at 6D9/DEAF7B60
	r = regexp.MustCompile(`(?s)logical decoding found consistent point at.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "logical_decoding_consistent_point"
		return result
	}

	//consistent recovery state reached at 12C/552375D0
	r = regexp.MustCompile(`(?s)consistent recovery state reached at.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "consistent_recovery_state_reached"
		return result
	}

	//starting logical decoding for slot...
	r = regexp.MustCompile(`(?s)starting logical decoding for slot.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "starting_logical_decoding"
		return result
	}

	//autovacuum launcher started
	r = regexp.MustCompile(`(?s)autovacuum launcher started.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "autovacuum_launcher_started"
		return result
	}

	//database system is shut down
	r = regexp.MustCompile(`(?s)database system (is|was) shut down.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "database_shut_down"
		return result
	}

	//database system is ready to accept read only connections
	r = regexp.MustCompile(`(?s)database system (is|was) shut down.*`)
	match = r.FindStringSubmatch(message)

	if len(match) > 0 {
		result["logType"] = "standby_is_up"
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
			q.minDuration = duration
			q.maxDuration = duration
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
	// weekday
	weekday := q.timestamp.Weekday()
	b, err := json.Marshal(fmt.Sprintf("%s", weekday))
	if err != nil {
		return nil, err
	}

	weekdayStr := json.RawMessage(b)
	q.data["day_of_week"] = &weekdayStr

	bInt, err := json.Marshal(int(weekday))
	if err != nil {
		return nil, err
	}
	weekdayInt := json.RawMessage(bInt)
	q.data["day_of_week_int"] = &weekdayInt

	// count
	b, err = json.Marshal(q.totalCount)
	if err != nil {
		return nil, err
	}
	rawCount := json.RawMessage(b)
	q.data["total_count"] = &rawCount

	// minDuration
	b, err = json.Marshal(q.minDuration)
	if err != nil {
		return nil, err
	}
	minDuration := json.RawMessage(b)
	q.data["min_duration_ms"] = &minDuration

	// maxDuration
	b, err = json.Marshal(q.maxDuration)
	if err != nil {
		return nil, err
	}
	maxDuration := json.RawMessage(b)
	q.data["max_duration_ms"] = &maxDuration

	// duration rounded to 5 decimal points
	b, err = json.Marshal(round(q.totalDuration, 0.5, 5))
	if err != nil {
		return nil, err
	}
	rawDuration := json.RawMessage(b)
	q.data["total_duration_ms"] = &rawDuration

	// avg duration rounded to 5 decimal points
	b, err = json.Marshal(round((q.totalDuration / float64(q.totalCount)), 0.5, 5))
	if err != nil {
		return nil, err
	}
	rawAvgDuration := json.RawMessage(b)
	q.data["avg_duration_ms"] = &rawAvgDuration

	// codeSourceMap to unique slice
	for _, v := range q.codeSourceMap {
		q.codeSource = append(q.codeSource, v)
	}

	b, err = json.Marshal(q.codeSource)
	if err != nil {
		return nil, err
	}
	codeSource := json.RawMessage(b)
	q.data["code_source"] = &codeSource

	// code application from comments
	if q.codeApplication != "" {
		b, err = json.Marshal(q.codeApplication)
		if err != nil {
			return nil, err
		}
		codeApplication := json.RawMessage(b)
		q.data["code_application"] = &codeApplication
	}

	// code controller from comments
	if q.codeController != "" {
		b, err = json.Marshal(q.codeController)
		if err != nil {
			return nil, err
		}
		codeController := json.RawMessage(b)
		q.data["code_controller"] = &codeController
	}

	// code job from comments
	if q.codeJob != "" {
		b, err = json.Marshal(q.codeJob)
		if err != nil {
			return nil, err
		}
		codeJob := json.RawMessage(b)
		q.data["code_job"] = &codeJob
	}

	// code line from comments
	if q.codeLine != "" {
		b, err = json.Marshal(q.codeLine)
		if err != nil {
			return nil, err
		}
		codeLine := json.RawMessage(b)
		q.data["code_line"] = &codeLine
	}

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

	// redisKey
	if q.redisKey != "" {
		err = marshalString(q, q.redisKey, "redis_key")
		if err != nil {
			return nil, err
		}
	}

	// add value if commandTag is blank
	if q.commandTag == "" {
		err = marshalString(q, "UNKNOWN", "command_tag")
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

		batchMap[batch{roundMin, q.uniqueSha}].codeSource = append(batchMap[batch{roundMin, q.uniqueSha}].codeSource, q.codeSource...)

		// caclulate min/max duration
		if batchMap[batch{roundMin, q.uniqueSha}].minDuration > q.minDuration {
			batchMap[batch{roundMin, q.uniqueSha}].minDuration = q.minDuration
		}
		if batchMap[batch{roundMin, q.uniqueSha}].maxDuration < q.maxDuration {
			batchMap[batch{roundMin, q.uniqueSha}].maxDuration = q.maxDuration
		}
	} else {
		batchMap[batch{roundMin, q.uniqueSha}] = q
	}
	mutex.Unlock()
}

func iterOverQueries() {
	var (
		duration time.Duration
		count    int64
	)
	now := currentMinute()
	mutex.Lock()
	for k := range batchMap {
		duration = now.Sub(k.minute)
		if duration >= (1 * time.Minute) {
			count++
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
	if count > 0 {
		color.Set(color.FgCyan)
		logit.Info(" Sent %d messages to ES Bulk Processor", count)
		color.Unset()
	}
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
