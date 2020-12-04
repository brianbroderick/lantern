package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestExtractDetails(t *testing.T) {
	initialSetup()
	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("extract.json")
	conn.Do("LPUSH", redisKey(), sample)

	uniq := "select c0.\"id\" from \"some_table\" as c0 where (c0.\"name\" ilike $1) and (c0.\"location_uid\" = $2) and (c0.\"user_uid\" = $3) limit $4"
	query, err := getLog(redisKey())

	assert.NoError(t, err)
	assert.Equal(t, uniq, query.uniqueStr)

	qd := new(queryDetails)
	qd.fragment = "ilike"
	qd.columns = []string{"location_uid", "user_uid"}

	assert.True(t, query.matchFragment(qd))

	query.extractDetails()
	assert.Equal(t, "721e69b2-af3d-52f8-a2a6-af630baa4be8", query.detailMap["$2"])

	query.findParam(qd)

	assert.Equal(t, map[string]string{"location_uid": "$2", "user_uid": "$3"}, query.paramMap)

	query.resolveParams()

	assert.Equal(t, map[string]string{
		"location_uid": "721e69b2-af3d-52f8-a2a6-af630baa4be8",
		"user_uid":     "d0aff49b-5feb-5c47-9408-56491615682f"},
		query.resolvedParams)
}
