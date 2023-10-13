package deprecated

import (
	"os"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

// func TestRegEx(t *testing.T) {
// 	query := "select * from servers where id IN (?, ?, ?) and name = ?"
// 	expected := "select * from servers where id IN (?) and name = ?"

// 	assert.Equal(t, expected, truncateInLists(query))
// }
