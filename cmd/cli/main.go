package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/maya-florenko/spotis/internal/app"
	"github.com/maya-florenko/spotis/internal/banner"
	"github.com/maya-florenko/spotis/internal/spotify"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	if err := spotify.Init(ctx); err != nil {
		log.Fatal(err)
	}

	banner.Print()

	if err := app.Init(ctx); err != nil {
		log.Fatal(err)
	}
}
