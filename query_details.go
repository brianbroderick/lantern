package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type queryDetails struct {
	fragment string
	columns  []string
}

func (q *query) matchFragment(qd *queryDetails) bool {
	return strings.Contains(q.uniqueStr, qd.fragment)
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

func (q *query) findParam(qd *queryDetails) {
	q.paramMap = make(map[string]string)

	for _, c := range qd.columns {
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

func (q *query) resolveParams() {
	q.resolvedParams = make(map[string]string)
	for k, v := range q.paramMap {
		q.resolvedParams[k] = q.detailMap[v]
	}
}
