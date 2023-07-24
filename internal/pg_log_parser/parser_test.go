package pgLogParser

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

// Tests that we can parse a duration and return it to string form
func TestDuration(t *testing.T) {
	duration, _ := time.ParseDuration("0.059ms")
	assert.Equal(t, "0.059ms", fmt.Sprintf("%.3fms", float32(duration.Microseconds())/float32(1000)))
}

func TestParser(t *testing.T) {
	s := "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from foo where bar = $1"
	stmt, err := NewParser(strings.NewReader(s)).ParseStatement()
	assert.NoError(t, err)
	assert.Equal(t, "2023-07-10 09:52:46 MDT: select * from foo where bar = $1", stmt.ShortString())
}

func TestParserParseStatement(t *testing.T) {
	var tests = []struct {
		s   string
		obj Statement
		p   string
		err string
	}{
		// Single log entry without query
		{
			s: `2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from users where id = $1`,
			obj: &LogStatement{
				date:            "2023-07-10",
				time:            "09:52:46",
				timezone:        "MDT",
				remoteHost:      "127.0.0.1",
				remotePort:      50032,
				user:            "postgres",
				database:        "sampledb",
				pid:             24649,
				severity:        "LOG",
				durationLit:     "0.059",
				durationMeasure: "ms",
				preparedStep:    "execute",
				preparedName:    "unnamed",
				statement:       "select * from users where id = $1",
			},
			p: `2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from users where id = $1`,
		},
	}

	for _, tt := range tests {
		obj, err := NewParser(strings.NewReader(tt.s)).ParseStatement()
		assert.NoError(t, err)
		assert.Equal(t, tt.obj, obj)
		assert.Equal(t, tt.p, tt.obj.String())
	}
}
