package main

import (
	"os"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}
