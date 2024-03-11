package repo

import "encoding/json"

func MarshalJSON(data interface{}) string {
	b, err := json.MarshalIndent(data, "", "  ")
	if HasErr("marshallJSON", err) {
		return ""
	}
	return string(b)
}

func UnmarshalJSON(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	HasErr("unmarshallJSON", err)
}
