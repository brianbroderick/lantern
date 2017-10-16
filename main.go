package main

var (
	//RedisKey is the key that postgres logs to
	RedisKey = "postgres"
)

func main() {
	initialSetup()

	forever := make(chan bool)
	<-forever
}

func initialSetup() {
	SetupRedis()
}
