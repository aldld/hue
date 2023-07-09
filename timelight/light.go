package timelight

import (
	"fmt"
	"time"

	"github.com/aldld/hue/hue"
	"golang.org/x/exp/slog"
)

const (
	MinMirek      = 153
	MaxMirek      = 500
	MinBrightness = 0
	MaxBrightness = 100
)

type LightID string

type Light struct {
	h *hue.Client

	ID                  LightID
	Active              bool // Is Timelight currently controlling this light?
	HasColor            bool
	HasColorTemperature bool
	HasBrightness       bool

	LastUpdated time.Time
	TargetState TargetState
}

func (l *Light) SetActive() {
	l.Active = true
}

func (l *Light) SetInactive() {
	l.Active = false
}

func (l *Light) restrictTarget(target TargetState) TargetState {
	if !l.HasBrightness {
		target.HasBrightness = false
		target.Brightness = 0
	}
	if !l.HasColorTemperature {
		target.HasTempMirek = false
		target.TempMirek = 0
	}
	return target
}

func (l *Light) Update(now time.Time, target TargetState, duration time.Duration) error {
	target = l.restrictTarget(target)

	if l.TargetState == target {
		// New target state equals last target state, nothing to do.
		return nil
	}

	var update hue.LightUpdate
	if target.HasBrightness {
		update.Dimming = &hue.DimmingUpdate{
			Brightness: target.Brightness,
		}
	}
	if target.HasTempMirek {
		update.ColorTemperature = &hue.ColorTemperatureUpdate{
			Mirek: target.TempMirek,
		}
	}
	update.Dynamics = &hue.Dynamics{DurationMs: int(duration.Milliseconds())}

	if err := l.h.UpdateLight(string(l.ID), update); err != nil {
		return err
	}

	l.LastUpdated = now
	l.TargetState = target

	return nil
}

type TargetState struct {
	HasBrightness bool
	Brightness    float64 // 0 to 100

	HasTempMirek bool
	TempMirek    int // 153 to 500
}

var DefaultTargetState = TargetState{}

func (s TargetState) LogValue() slog.Value {
	brightness := "N/A"
	if s.HasBrightness {
		brightness = fmt.Sprintf("%v", s.Brightness)
	}
	tempMirek := "N/A"
	if s.HasTempMirek {
		tempMirek = fmt.Sprintf("%v", s.TempMirek)
	}
	return slog.GroupValue(
		slog.String("brightness", brightness),
		slog.String("temp_mirek", tempMirek),
	)
}

func (s TargetState) WithBrightness(brightness float64) TargetState {
	s.Brightness = brightness
	s.HasBrightness = true
	return s
}

func (s TargetState) WithColorTemp(mirek int) TargetState {
	s.TempMirek = mirek
	s.HasTempMirek = true
	return s
}

func (t *Timelight) initLights() error {
	t.lights = make(map[LightID]*Light)

	lights, err := t.hue.GetLights()
	if err != nil {
		return err
	}

	for _, l := range lights {
		light := &Light{
			h:      t.hue,
			ID:     LightID(l.ID),
			Active: false,

			HasColor:            l.Color != nil,
			HasColorTemperature: l.ColorTemperature != nil,
			HasBrightness:       l.Dimming != nil,
		}
		t.lights[light.ID] = light
	}

	t.log.Info("Initialized lights", slog.Int("count", len(t.lights)))

	return nil
}

func (t *Timelight) updateLights(now time.Time, target TargetState) {
	t.log.Info("updating lights", slog.Any("target", target))

	successes := 0
	errs := 0

	for _, light := range t.lights {
		if !light.Active {
			continue
		}

		err := light.Update(now, target, 10*time.Second)
		if err != nil {
			t.log.Error("error while updating light",
				slog.String("id", string(light.ID)),
				slog.Any("err", err),
			)
			errs += 1
		} else {
			successes += 1
		}
	}
	t.log.Info("finished updating lights",
		slog.Int("successes", successes),
		slog.Int("errs", errs),
	)
}
