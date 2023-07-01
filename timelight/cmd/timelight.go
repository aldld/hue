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
	var config timelight.Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		panic(err)
	}

	log := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      config.Logger.SlogLevel(),
		TimeFormat: time.TimeOnly,
	}))

	tl := timelight.New(log, config)
	if err := tl.Run(); err != nil {
		panic(err)
	}
}
