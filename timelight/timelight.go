package timelight

import (
	"golang.org/x/exp/slog"
	"time"

	"github.com/aldld/hue/hue"
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

	hue         *hue.Client
	lastEventId string

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

	if event.Type != "update" {
		return false
	}

	for _, r := range event.Data {
		if r.Type() == hue.RTypeScene || r.Type() == hue.RTypeLight {
			return true
		}
	}

	return false
}

func (t *Timelight) Run() error {
	t.log.Info("Starting Timelight")

	spec, err := t.config.Timelight.Spec()
	if err != nil {
		return err
	}

	// Initialize state. Query for scenes, identify lights to track.
	if err := t.initLights(); err != nil {
		return err
	}
	if err := t.initScenes(); err != nil {
		return err
	}

	bridgeEvents := make(chan hue.Event, 8)
	go t.hue.EventListener(filterEvent, bridgeEvents)

	t.runLightUpdate(time.Now(), spec)

	lightUpdate := time.Tick(lightUpdateInterval)
	for {
		select {
		case event := <-bridgeEvents:
			t.handleEvent(event)

		case <-lightUpdate:
			t.runLightUpdate(time.Now(), spec)
		}
	}
}

func (t *Timelight) runLightUpdate(now time.Time, spec Spec) {
	target := spec.TargetLightState(now)
	t.updateLights(now, target)
	t.updateScenes(target)
}
