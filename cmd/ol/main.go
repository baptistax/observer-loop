package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/baptistax/observer-loop/internal/watcher"
)

func main() {
	cfg, err := parseFlags()
	if err != nil {
		fail(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := watcher.Run(ctx, cfg); err != nil && !errors.Is(err, context.Canceled) {
		fail(err)
	}
}

func parseFlags() (watcher.Config, error) {
	var cfg watcher.Config

	flag.UintVar(&cfg.PID, "pid", 0, "PID to watch first")
	flag.StringVar(&cfg.Title, "title", "Observer Loop", "message box title")
	cfg.Out = os.Stdout
	flag.Parse()

	if cfg.PID == 0 {
		return watcher.Config{}, fmt.Errorf("usage: %s -pid <PID> [-title \"Watch Demo\"]", os.Args[0])
	}

	return cfg, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
