package main

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestParamsStruct(t *testing.T) {
	initialSetup()

	conn := pool.Get()
	defer conn.Close()

	qd := new(queryDetails)
	qd.fragment = "ilike"

	qd.columns = make(map[string]map[string]column)
	qd.columns["location_uid"] = make(map[string]column)
	qd.columns["user_uid"] = make(map[string]column)

	sample := readPayload("extract.json")
	conn.Do("LPUSH", redisKey(), sample)

	llen, err := conn.Do("LLEN", redisKey())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), llen)

	uniq := "select c0.\"id\" from \"some_table\" as c0 where (c0.\"name\" ilike $1) and (c0.\"location_uid\" = $2) and (c0.\"user_uid\" = $3) limit $4"
	comments := []string(nil)
	query, err := getLog(redisKey())

	assert.NoError(t, err)
	assert.Equal(t, uniq, query.uniqueStr)
	assert.Equal(t, comments, query.comments)

	// Does the query contain "ilike"?
	if strings.Contains(query.uniqueStr, qd.fragment) {
		result := make(map[string]string)

		r := regexp.MustCompile(`"location_uid" = (?P<param>\$\d+)`)
		match := r.FindStringSubmatch(query.uniqueStr)

		if len(match) > 0 {
			for i, name := range r.SubexpNames() {
				if i != 0 {
					result[name] = match[i]
				}
			}
		}

		assert.Equal(t, "$2", result["param"])

		query.detail = strings.ReplaceAll(query.detail, "parameters:", "")
		details := strings.Split(query.detail, ",")
		re := regexp.MustCompile(`(?P<param>\$\d+) = '(?P<value>.*)'`)
		paramMap := make(map[string]string)

		for _, d := range details {
			detailsMap := make(map[string]string)
			match = re.FindStringSubmatch(d)
			if len(match) > 0 {
				for i, name := range re.SubexpNames() {
					if i != 0 {
						detailsMap[name] = match[i]
					}
				}
			}

			paramMap[detailsMap["param"]] = detailsMap["value"]
		}

		assert.Equal(t, "721e69b2-af3d-52f8-a2a6-af630baa4be8", paramMap["$2"])
	}

}
