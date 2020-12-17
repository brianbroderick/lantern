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

	assert.Equal(t, "location_uid", batchDetailsMap[batch{minute, "0db40a64f409661d773d52075f4cd00531aee122"}].column)
	assert.Equal(t, "721e69b2-af3d-52f8-a2a6-af630baa4be8", batchDetailsMap[batch{minute, "0db40a64f409661d773d52075f4cd00531aee122"}].columnValue)

	assert.Equal(t, "user_uid", batchDetailsMap[batch{minute, "1adf948179710ba33ac5ed636660f9335b6a250b"}].column)
	assert.Equal(t, "d0aff49b-5feb-5c47-9408-56491615682f", batchDetailsMap[batch{minute, "1adf948179710ba33ac5ed636660f9335b6a250b"}].columnValue)

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

	assert.Equal(t, "location_uid", batchDetailsMap[batch{minute, "0db40a64f409661d773d52075f4cd00531aee122"}].column)
	assert.Equal(t, "8819196b-d5b4-47d5-9646-8645e2b2ed85", batchDetailsMap[batch{minute, "0db40a64f409661d773d52075f4cd00531aee122"}].columnValue)

	assert.Equal(t, "user_uid", batchDetailsMap[batch{minute, "1adf948179710ba33ac5ed636660f9335b6a250b"}].column)
	assert.Equal(t, "b15e38eb-c09f-46b7-ae35-deee2cdffad2", batchDetailsMap[batch{minute, "1adf948179710ba33ac5ed636660f9335b6a250b"}].columnValue)

}
