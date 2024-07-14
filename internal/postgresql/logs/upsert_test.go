package logs

import (
	"fmt"
	"testing"
	"time"
)

func TestUpsertData(t *testing.T) {
	t1 := time.Now()
	UpsertQueries()
	UpsertDatabases()
	ExtractAndUpsertQueryMetadata()
	t2 := time.Now()
	timeDiff := t2.Sub(t1)
	fmt.Printf("\nTime Elapsed: %v\n", timeDiff)
}
