package logs

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAggregateLogs(t *testing.T) {
	databases, queries := AggregateLogs("TestAggregateLogs", SampleCreateLog(), "queries-test.json", "databases-test.json")

	assert.Equal(t, 1, len(databases.Databases), "Number of databases")
	assert.Equal(t, 7, len(queries.Queries), "Number of queries")

	// for _, query := range queries.Queries {
	// 	assert.Equal(t, "2024-07-10 17:00:00 UTC", query.TimestampByHour.UTC().Format("2006-01-02 15:04:05 MST"), "TimestampByHour")
	// }
}

func TestTimeZone(t *testing.T) {
	hourFormat := "2006-01-02 15:00:00 MST"

	timestamp, err := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s %s", "1975-06-19", "23:12:19"), loadTz("MDT"))
	assert.NoError(t, err, "Time Parse")

	hourstamp := timestamp.UTC().Format(hourFormat)
	assert.Equal(t, "1975-06-20 05:00:00 UTC", hourstamp, "Hourstamp")
}

func SampleCreateLog() string {
	return `2024-07-10 17:48:11 UTC:10.0.0.1(59454):myuser@lantern:[44600]:LOG:  duration: 0.142 ms  statement: DISCARD ALL;
2024-07-10 17:48:11 UTC:10.0.0.1(48684):myuser@lantern:[40113]:LOG:  statement: set statement_timeout = '360s'; /*{"somekey":42, "another-key": "some value"}*/drop table if exists temp_tbl;create temp table temp_tbl as ( select
		uid, address_id, name
	from
		details
	where true and address_id = 1  );create index idx_temp_tbl_address_id on temp_tbl using btree( uid, address_id );analyze temp_tbl;	
2024-07-10 17:53:12 UTC:10.0.0.1(48684):myuser@lantern:[40113]:LOG:  duration: 27.176 ms`
}
