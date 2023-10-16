package timelight

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/slog"
)

type Config struct {
	Logger    LoggerConfig    `toml:"logger"`
	Bridge    BridgeConfig    `toml:"bridge"`
	Timelight TimelightConfig `toml:"timelight"`
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

type StateConfig struct {
	Brightness     *float64 `toml:"brightness,omitempty"`
	ColorTempMirek *int     `toml:"color_temp_mirek,omitempty"`
}

func (c StateConfig) TargetState() TargetState {
	var state TargetState

	if c.Brightness != nil {
		b := *c.Brightness

		var brightness float64
		if b < MinBrightness {
			brightness = MinBrightness
		} else if b > MaxBrightness {
			brightness = MaxBrightness
		} else {
			brightness = b
		}
		state = state.WithBrightness(brightness)
	}

	if c.ColorTempMirek != nil {
		t := *c.ColorTempMirek

		var temp int
		if t < MinMirek {
			temp = MinMirek
		} else if t > MaxMirek {
			temp = MaxMirek
		} else {
			temp = t
		}

		state = state.WithColorTemp(temp)
	}

	return state
}

type TimelightConfig struct {
	Brightness TransitionConfig `toml:"brightness"`
	ColorTemp  TransitionConfig `toml:"color_temp"`
}

func (c TimelightConfig) Spec() (Spec, error) {
	brightness, err := c.Brightness.spec()
	if err != nil {
		return nil, err
	}

	temp, err := c.ColorTemp.spec()
	if err != nil {
		return nil, err
	}

	return SmoothSpec{Brightness: brightness, TempMirek: temp}, nil
}

type TransitionConfig struct {
	StartTime      string `toml:"start_time"`
	TransitionTime string `toml:"transition_time"`
	EndTime        string `toml:"end_time"`

	StartValue int `toml:"start_value"`
	EndValue   int `toml:"end_value"`
}

func (c TransitionConfig) spec() (SmoothTransition, error) {
	var empty SmoothTransition

	startTime, err := parseMinuteOfDay(c.StartTime)
	if err != nil {
		return empty, err
	}

	transitionTime, err := parseMinuteOfDay(c.TransitionTime)
	if err != nil {
		return empty, err
	}

	endTime, err := parseMinuteOfDay(c.EndTime)
	if err != nil {
		return empty, err
	}

	return SmoothTransition{
		StartMinute:      startTime,
		TransitionMinute: transitionTime,
		EndMinute:        endTime,

		Start: float64(c.StartValue),
		End:   float64(c.EndValue),
	}, nil
}

func parseMinuteOfDay(s string) (int, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time: %s", s)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	return 60*hour + minute, nil
}
