package extractor

import (
	"github.com/google/uuid"
)

// UuidNamespace is the namespace for the uuid. There are 4 predefined namespaces, but you can also create your own.
var UuidNamespace = uuid.MustParse("018e1b50-ee98-73f2-9839-420223323163")

func UuidV5(str string) uuid.UUID {
	return uuid.NewSHA1(UuidNamespace, []byte(str))
}

func UuidString(query string) string {
	return UuidV5(query).String()
}
