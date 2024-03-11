package repo

import (
	"crypto/sha1"
	"encoding/hex"
	"io"

	"github.com/google/uuid"
)

// UuidNamespace is the namespace for the uuid. There are 4 predefined namespaces, but you can also create your own.
var UuidNamespace = uuid.MustParse("018e1b50-ee98-73f2-9839-420223323163")

// ShaQuery creates a sha of the query
func ShaQuery(query string) string {
	h := sha1.New()
	io.WriteString(h, query)

	return hex.EncodeToString(h.Sum(nil))
}

func UuidQuery(query string) uuid.UUID {
	return uuid.NewSHA1(UuidNamespace, []byte(query))
}
