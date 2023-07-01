package timelight

import (
	"strings"

	"golang.org/x/exp/slog"
)

type Config struct {
	Logger LoggerConfig `toml:"logger"`
	Bridge BridgeConfig `toml:"bridge"`
}

type LoggerConfig struct {
	Level string `toml:"level"`
}

func (c LoggerConfig) SlogLevel() slog.Level {
	switch strings.ToLower(c.Level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type BridgeConfig struct {
	Addr     string `toml:"addr"`
	Username string `toml:"username"`
}
