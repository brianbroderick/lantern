package analyzer

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestParseToJSON(t *testing.T) {
	json := ParseToJSON("select 42;")
	assert.NotEmpty(t, json)

	fmt.Println(json)
}
