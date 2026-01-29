package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"pvz-service/internal/app/api"
	"pvz-service/internal/config"
)

func main() {
	cfg := config.Load()

	app, err := api.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		_ = app.Shutdown(context.Background())
	}()

	log.Fatal(app.Run(context.Background()))
}
