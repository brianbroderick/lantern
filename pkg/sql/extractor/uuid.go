package extractor

import (
	"github.com/google/uuid"
)

// UuidNamespace is the namespace for the uuid. There are 4 predefined namespaces, but you can also create your own.
var UuidNamespace = uuid.MustParse("018e1b50-ee98-73f2-9839-420223323163")

// Using the above UuidNamespace, we passed "default" into UuidString() and got the UUID "14734a7f-e995-5c6a-9af1-5cd82ccd1628"
var DefaultUUIDStr = "14734a7f-e995-5c6a-9af1-5cd82ccd1628"
var DefaultUUID = uuid.MustParse(DefaultUUIDStr)

func UuidV5(str string) uuid.UUID {
	return uuid.NewSHA1(UuidNamespace, []byte(str))
}

// UuidString returns a string representation of a UUIDv5
// For example, the UUID for "default" is "14734a7f-e995-5c6a-9af1-5cd82ccd1628"

func UuidString(query string) string {
	return UuidV5(query).String()
}
