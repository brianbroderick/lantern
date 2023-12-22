package evaluator

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
)

// ShaQuery creates a sha of the query
func ShaQuery(query string) string {
	h := sha1.New()
	io.WriteString(h, query)

	return hex.EncodeToString(h.Sum(nil))
}
