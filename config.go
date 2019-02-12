package main

import (
	"flag"
	"time"
)

var (
	port        = flag.Int("port", 8000, "port for http server")
	memoryLimit = flag.Int(
		"limit", 100, "maximum number of messages which can be hold in memory",
	)
	senderPeriod      = flag.Duration("period", 1*time.Second, "message sender period")
	messageBirdSecret = flag.String("api-token", "", "message bird api token")
)

type Config struct {
	ServerPort int

	SenderMemoryLimit int
	SenderPeriod      time.Duration
	SenderApiToken    string
}

func GetConfig() Config {
	flag.Parse()

	return Config{
		ServerPort:        *port,
		SenderMemoryLimit: *memoryLimit,
		SenderPeriod:      *senderPeriod,
		SenderApiToken:    *messageBirdSecret,
	}
}
