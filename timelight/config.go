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
	Transition     string `toml:"transition"`
	StartTime      string `toml:"start_time"`
	TransitionTime string `toml:"transition_time"`
	EndTime        string `toml:"end_time"`

	Start StateConfig `toml:"start"`
	End   StateConfig `toml:"end"`
}

func (c TimelightConfig) Spec() (Spec, error) {
	start := c.Start.TargetState()
	end := c.End.TargetState()

	switch c.Transition {
	case "smooth":
		startTime, err := parseMinuteOfDay(c.StartTime)
		if err != nil {
			return nil, err
		}

		transitionTime, err := parseMinuteOfDay(c.TransitionTime)
		if err != nil {
			return nil, err
		}

		endTime, err := parseMinuteOfDay(c.EndTime)
		if err != nil {
			return nil, err
		}

		return SmoothSpec{
			StartMinute:      startTime,
			TransitionMinute: transitionTime,
			EndMinute:        endTime,

			Start: start,
			End:   end,
		}, nil

	case "linear":
		startTime, err := parseMinuteOfDay(c.StartTime)
		if err != nil {
			return nil, err
		}

		endTime, err := parseMinuteOfDay(c.EndTime)
		if err != nil {
			return nil, err
		}

		return LinearSpec{
			StartMinute: startTime,
			EndMinute:   endTime,

			Start: start,
			End:   end,
		}, nil
	default:
		return nil, fmt.Errorf("unknown transition: %s", c.Transition)
	}
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
