package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// for the details field, the first map is each column in question
// the second map key is the column value
type queryDetails struct {
	fragment string
	columns  []string
}

type queryDetailStats struct {
	uid           string  // sha of uniqueSha, column, and columnValue
	uniqueSha     string  // sha of uniqueStr and preparedStep (if available). This matches what's in the query struct
	column        string  // column in question such as "user_id"
	columnValue   string  // value of column such as "42"
	totalCount    int32   // number of times the column appeared in queries
	totalDuration float64 // total duration of each time the column appeared in queries
}

func (q *query) handleQueryDetails() {
	// If fragment was not found, return
	if !q.matchFragment() {
		return
	}

	q.extractDetails()
	q.findParam()
	q.addToDetails()
}

func newQueryDetails(fragment string, columns string) *queryDetails {
	qd := new(queryDetails)

	if columns == "" {
		return qd
	}

	qd.fragment = fragment
	qd.columns = strings.Split(columns, ",")

	return qd
}

func (q *query) matchFragment() bool {
	return strings.Contains(q.uniqueStr, detailArgs.fragment)
}

// q.detail looks like:
// "parameters: $1 = '%brian%', $2 = '721e69b2-af3d-52f8-a2a6-af630baa4be8', $3 = 'd0aff49b-5feb-5c47-9408-56491615682f'"
func (q *query) extractDetails() {
	if source, pres := q.data["detail"]; pres {
		if err := json.Unmarshal(*source, &q.detail); err != nil {
			return
		}
	}

	// Guard against empty field
	if q.detail == "" {
		return
	}

	q.detail = strings.ReplaceAll(q.detail, "parameters:", "")

	details := strings.Split(q.detail, ",")

	re := regexp.MustCompile(`(?P<param>\$\d+) = '(?P<value>.*)'`)
	q.detailMap = make(map[string]string)

	for _, d := range details {
		match := re.FindStringSubmatch(d)
		if len(match) >= 2 {
			q.detailMap[match[1]] = match[2]
		}
	}
}

func (q *query) findParam() {
	q.paramMap = make(map[string]string)

	for _, c := range detailArgs.columns {
		// Match the following patterns. Column prefixes shouldn't affect this pattern.
		//   "user_uid" = $1
		//   user_uid = $1
		pattern := fmt.Sprintf(`"*%s"* = (?P<param>\$\d+)`, c)
		r := regexp.MustCompile(pattern)
		match := r.FindStringSubmatch(q.uniqueStr)

		if len(match) > 0 {
			q.paramMap[c] = match[1]
		}
	}
}

func (q *query) addToDetails() {
	minute := roundToMinute(q.timestamp)

	for k, v := range q.paramMap {
		uid := q.shaQueryDetailStats(k, v)

		// Multiple goroutines will access this hash
		detailsMutex.Lock()

		if _, ok := batchDetailsMap[batch{minute, uid}]; ok == true {
			batchDetailsMap[batch{minute, uid}].totalCount++
			batchDetailsMap[batch{minute, uid}].totalDuration += q.totalDuration
		} else {
			batchDetailsMap[batch{minute, uid}] = &queryDetailStats{
				uid:           uid,
				uniqueSha:     q.uniqueSha,
				column:        k,
				columnValue:   q.detailMap[v],
				totalCount:    1,
				totalDuration: q.totalDuration,
			}
		}

		detailsMutex.Unlock()
	}
}

func (q *query) shaQueryDetailStats(column string, columnValue string) string {
	h := sha1.New()
	io.WriteString(h, column)
	io.WriteString(h, columnValue)
	io.WriteString(h, q.uniqueSha)
	return hex.EncodeToString(h.Sum(nil))
}
