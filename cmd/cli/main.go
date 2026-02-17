package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/maya-florenko/spotis/internal/app"
	"github.com/maya-florenko/spotis/internal/banner"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	banner.Print()

	if err := app.Init(ctx); err != nil {
		log.Fatal(err)
	}
}