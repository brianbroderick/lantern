package logit

import (
	"fmt"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	t1 := time.Now()

	for i := 0; i < 100000; i++ {
		Append(fmt.Sprintf("%d: Hello World", i))
	}

	t2 := time.Now()
	timeDiff := t2.Sub(t1)
	avg := timeDiff / time.Duration(100000)
	fmt.Printf("TestLog, Elapsed Time: %s, Avg per line: %s\n", timeDiff, avg)
}
