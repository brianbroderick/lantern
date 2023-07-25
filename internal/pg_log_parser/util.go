package pgLogParser

import (
	"fmt"
)

// func readPayload(file string) ([]byte, error) {
// 	if len(file) == 0 {
// 		return []byte{}, errors.New("file is empty")
// 	}

// 	data, err := os.ReadFile(string(file))
// 	if hasErr(err) {
// 		return []byte{}, err
// 	}
// 	return data, nil
// }

func hasErr(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
		return true
	}
	return false
}
