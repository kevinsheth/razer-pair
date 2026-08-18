package main

import (
	"context"
	"os"
	"os/signal"

	"razer-pair/internal/cli"
	"razer-pair/internal/hidapi"
)

var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(cli.Run(ctx, os.Args[1:], cli.Options{
		Version:      version,
		RealProvider: hidapi.NewProvider(),
		Stdin:        os.Stdin,
		Stdout:       os.Stdout,
		Stderr:       os.Stderr,
	}))
}
