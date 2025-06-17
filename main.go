package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/Leul-Michael/image-generation/application"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

func main() {
	app, err := application.New()
	if err != nil {
		log.Fatal("failed to initialize the app: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err = app.Start(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
