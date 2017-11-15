package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestElasticURL(t *testing.T) {
	origURL := os.Getenv("ELASTIC_URL")
	os.Setenv("ELASTIC_URL", "")

	elasticURL := elasticURL()
	assert.Equal(t, "http://127.0.0.1:9200", elasticURL)

	os.Setenv("ELASTIC_URL", origURL)
}
