package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/messagebird/go-rest-api"
)

func main() {
	cfg := GetConfig()

	client := &messageBirdClientWrapper{messagebird.New(cfg.SenderApiToken)}
	sender := NewSender(
		cfg.SenderMemoryLimit,
		cfg.SenderPeriod,
		client,
	)

	server := NewServer(cfg.ServerPort, sender.Enque)
	server.Start()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	<-exit

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint
	server.Close(ctx)
	sender.Close()
}
