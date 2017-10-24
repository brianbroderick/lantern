package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var queryMap = make(map[string]*query)

func main() {
	initialSetup()

	forever := make(chan bool)
	<-forever
}

func initialSetup() {
	setupEnv()
	SetupRedis()
}

func setupEnv() {
	platformEnv := os.Getenv("PLATFORM_ENV")
	if platformEnv != "prod" && platformEnv != "stage" {
		filename := ".env_" + platformEnv
		log.Printf("INFO: loading file %s", filename)
		err := godotenv.Load(filename)
		if err != nil {
			log.Printf("INFO: %s file not found", filename)
		}
	}
	log.Printf("INFO: loading file .env")
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("INFO: .env file not found")
	}
}

func redisKey() string {
	platformEnv := os.Getenv("PLATFORM_ENV")
	if platformEnv != "test" {
		return "postgres"
	}

	return "postgres_test"
}
