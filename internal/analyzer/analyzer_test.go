package analyzer

import (
	"os"
	"testing"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestAnalyze(t *testing.T) {
	Analyze()
}
