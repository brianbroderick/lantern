package main

import "io/ioutil"

func writeFile(prefix string, uniqueHash string, body []byte) {
	d1 := []byte(body)

	err := ioutil.WriteFile("./automated_payloads/"+prefix+"_"+uniqueHash, d1, 0644)
	check(err)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
