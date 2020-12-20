package main

import (
	"os"
	"testing"

	logit "github.com/brianbroderick/logit"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestHandleQueryDetailsEachStep(t *testing.T) {
	initialSetup()
	SetupElastic()
	truncateElasticSearch()

	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("extract.json")
	conn.Do("LPUSH", redisKey(), sample)

	uniq := "select c0.\"id\" from \"some_table\" as c0 where (c0.\"name\" ilike $1) and (c0.\"location_uid\" = $2) and (c0.\"user_uid\" = $3) limit $4"
	query, err := getLog(redisKey())

	assert.NoError(t, err)
	assert.Equal(t, uniq, query.uniqueStr)

	detailArgs = newQueryDetails("ilike", "location_uid,user_uid")

	// qd := newQueryDetails("ilike", []string{"location_uid", "user_uid"})

	assert.True(t, query.matchFragment())

	query.extractDetails()
	assert.Equal(t, "721e69b2-af3d-52f8-a2a6-af630baa4be8", query.detailMap["$2"])

	query.findParam()
	assert.Equal(t, map[string]string{"location_uid": "$2", "user_uid": "$3"}, query.paramMap)

	query.addToDetails()

	minute := roundToMinute(query.timestamp)

	assert.Equal(t, "location_uid", batchDetailsMap[batch{minute, "dcf527cc0d74e8aaca1e62212d99c1cdb44faf12"}].column)
	assert.Equal(t, "721e69b2-af3d-52f8-a2a6-af630baa4be8", batchDetailsMap[batch{minute, "dcf527cc0d74e8aaca1e62212d99c1cdb44faf12"}].columnValue)

	assert.Equal(t, "user_uid", batchDetailsMap[batch{minute, "3439739962158697084b622343411320966ef2bf"}].column)
	assert.Equal(t, "d0aff49b-5feb-5c47-9408-56491615682f", batchDetailsMap[batch{minute, "3439739962158697084b622343411320966ef2bf"}].columnValue)

	iterOverDetails()
	assert.Equal(t, 0, len(batchDetailsMap))

	err = bulkProc["bulk"].Flush()
	if err != nil {
		logit.Error("Error flushing messages: %e", err.Error())
	}

	conn.Do("DEL", redisKey())
}

func TestHandleQueryDetails(t *testing.T) {
	initialSetup()

	redisKey := "TestHandleQueryDetails"
	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("extract_query_details.json")
	conn.Do("LPUSH", redisKey, sample)

	detailArgs = newQueryDetails("ilike", "location_uid,user_uid")

	query, _ := getLog(redisKey)
	query.handleQueryDetails()

	minute := roundToMinute(query.timestamp)

	assert.Equal(t, "location_uid", batchDetailsMap[batch{minute, "9abbefa5a8e865f6b191b1124e5631c08fc43c71"}].column)
	assert.Equal(t, "8819196b-d5b4-47d5-9646-8645e2b2ed85", batchDetailsMap[batch{minute, "9abbefa5a8e865f6b191b1124e5631c08fc43c71"}].columnValue)

	assert.Equal(t, "user_uid", batchDetailsMap[batch{minute, "4c4fbcc9ad216febbc3e77e8b99c50ab0c37cda4"}].column)
	assert.Equal(t, "b15e38eb-c09f-46b7-ae35-deee2cdffad2", batchDetailsMap[batch{minute, "4c4fbcc9ad216febbc3e77e8b99c50ab0c37cda4"}].columnValue)
}

func TestHandleQueryDetailsWithCommas(t *testing.T) {
	initialSetup()

	redisKey := "TestHandleQueryDetails"
	conn := pool.Get()
	defer conn.Close()

	sample := readPayload("extract_with_commas.json")
	conn.Do("LPUSH", redisKey, sample)

	detailArgs = newQueryDetails("ilike", "location_uid,user_uid")

	query, _ := getLog(redisKey)
	query.handleQueryDetails()

	minute := roundToMinute(query.timestamp)

	assert.Equal(t, "location_uid", batchDetailsMap[batch{minute, "46a0a6742ce40bc5002314cdeaaa0f0eb88bc848"}].column)
	assert.Equal(t, "d5d3c4d6-735f-413a-a360-96acbe1d943b", batchDetailsMap[batch{minute, "46a0a6742ce40bc5002314cdeaaa0f0eb88bc848"}].columnValue)

	assert.Equal(t, "user_uid", batchDetailsMap[batch{minute, "51554bfe74c13945b63cb5729701ddbec78e6425"}].column)
	assert.Equal(t, "6f962f29-53dd-437f-a202-11f7d7679586", batchDetailsMap[batch{minute, "51554bfe74c13945b63cb5729701ddbec78e6425"}].columnValue)

}
