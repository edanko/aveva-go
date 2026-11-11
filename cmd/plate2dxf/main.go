package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"log/slog"

	"github.com/edanko/aveva-go/cmd/plate2dxf/internal/app"
)

// errInvalidMirror reports a bad -mirror value.
var errInvalidMirror = errors.New("invalid -mirror value")

var (
	inputDir  = flag.String("i", ".", "input directory")
	outputDir = flag.String("o", ".", "output directory")

	mirror = flag.String("mirror", "auto", `part orientation: "auto" (flip when the OS marking is longer), "on" (always flip), "off" (never flip)`)
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	mode := app.MirrorMode(*mirror)
	switch mode {
	case app.MirrorAuto, app.MirrorOn, app.MirrorOff:
	default:
		logger.Error("invalid -mirror value", "value", *mirror, "want", "auto|on|off")

		return errInvalidMirror
	}

	a := app.New(logger, app.Options{Mirror: mode})
	if err := a.Process(ctx, os.DirFS(*inputDir), *outputDir); err != nil {
		logger.Error("processing failed", "error", err)

		return err
	}

	return nil
}
