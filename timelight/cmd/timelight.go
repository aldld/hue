package main

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/lmittmann/tint"
	"golang.org/x/exp/slog"

	"github.com/aldld/hue/timelight"
)

func main() {
	args := os.Args
	configFilename := "config.toml"
	if len(args) >= 2 {
		configFilename = args[len(args)-1]
	}

	var config timelight.Config
	if _, err := toml.DecodeFile(configFilename, &config); err != nil {
		panic(err)
	}

	log := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      config.Logger.SlogLevel(),
		TimeFormat: time.TimeOnly,
	}))

	tl := timelight.New(log, config)
	if err := tl.Run(); err != nil {
		log.Error("timelight errored", slog.Any("err", err))
		os.Exit(1)
	}
}
