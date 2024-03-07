package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConn tests the database connection is valid and can be pinged
func TestConn(t *testing.T) {
	db := Conn()
	defer db.Close()

	assert.NoError(t, db.Ping())
}
