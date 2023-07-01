package timelight

import (
	"time"

	"github.com/aldld/hue/hue"
	"golang.org/x/exp/slog"
)

const (
	lightUpdateInterval = 1 * time.Minute
)

// Periodically checkpoint (every 1 minute?) the following data to disk:
//   - Last event ID from the event stream
//   - State of all lights being tracked
//   - Timestamp
// On startup, check if checkpoint file exists and timestamp is not older than ~2min(?)
// Restore state of lights being tracked, and resume stream with Last-Event-Id header

type Timelight struct {
	log    *slog.Logger
	config Config

	hue *hue.Client

	scenes map[SceneID]*Scene
	lights map[LightID]*Light
}

func New(log *slog.Logger, config Config) *Timelight {
	hueConfig := hue.Config{
		Addr:   config.Bridge.Addr,
		AppKey: config.Bridge.Username,
	}
	hueClient := hue.NewClient(log, hueConfig)

	return &Timelight{
		log:    log,
		config: config,
		hue:    hueClient,
	}
}

func filterEvent(event hue.Event) bool {
	// TODO: Events that we care about:
	// Light state changed
	// Scene added, modified, deleted
	// Scene recalled
	// TODO: How to filter out events caused by timelight?
	// Keep track of last updated time for each light/scene/resource
	// Compare with creationTime on event, ignore if within delta (~2s)
	return true
}

func (t *Timelight) Run() error {
	t.log.Info("Starting Timelight")

	// TODO make these configs
	spec := LinearSpec{
		StartMinute: 6 * 60,     // 06:00
		EndMinute:   23*60 + 59, // 23:59

		Start: DefaultTargetState.WithBrightness(100).WithColorTemp(MinMirek),
		End:   DefaultTargetState.WithBrightness(60).WithColorTemp(MaxMirek),
	}

	// Initialize state. Query for scenes, identify lights to track.
	if err := t.initLights(); err != nil {
		return err
	}
	if err := t.initScenes(); err != nil {
		return err
	}

	// TODO Figure out which lights should be considered active

	bridgeEvents := make(chan hue.Event, 8)
	go t.hue.EventListener(filterEvent, bridgeEvents)

	lightUpdate := time.Tick(lightUpdateInterval)
	for {
		select {
		case event := <-bridgeEvents:
			t.log.Info("Handling event", slog.Any("event", event))

		case <-lightUpdate:
			target := spec.TargetLightState(time.Now())
			t.updateLights(target)
			t.updateScenes(target)
		}
	}
}

// Spec defines a rule for computing the target brightness and color temperature
// for a given time.
type Spec interface {
	TargetLightState(now time.Time) TargetState
}

type LinearSpec struct {
	StartMinute int
	EndMinute   int

	Start TargetState
	End   TargetState
}

func (s LinearSpec) TargetLightState(now time.Time) TargetState {
	curMinute := minuteOfDay(now)
	if curMinute < s.StartMinute || curMinute > s.EndMinute {
		return s.End
	}

	p := s.progressPct(curMinute)

	var target TargetState
	if s.Start.HasBrightness && s.End.HasBrightness {
		target = target.WithBrightness(s.Start.Brightness + p*(s.End.Brightness-s.Start.Brightness))
	} else if s.Start.HasBrightness && !s.End.HasBrightness {
		target = target.WithBrightness(s.Start.Brightness)
	} else if !s.Start.HasBrightness && s.End.HasBrightness {
		target = target.WithBrightness(s.End.Brightness)
	}

	if s.Start.HasTempMirek && s.End.HasTempMirek {
		target = target.WithColorTemp(int(float64(s.Start.TempMirek) + p*float64(s.End.TempMirek-s.Start.TempMirek)))
	} else if s.Start.HasTempMirek && !s.End.HasTempMirek {
		target = target.WithColorTemp(s.Start.TempMirek)
	} else if !s.Start.HasTempMirek && s.End.HasTempMirek {
		target = target.WithColorTemp(s.End.TempMirek)
	}

	return target
}

func minuteOfDay(t time.Time) int {
	hour, min, _ := t.Clock()
	return 60*hour + min
}

func (s LinearSpec) progressPct(curMinute int) float64 {
	p := float64(curMinute-s.StartMinute) / float64(s.EndMinute-s.StartMinute)
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}
