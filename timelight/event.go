package timelight

import (
	"fmt"
	"math"
	"time"

	"github.com/aldld/hue/hue"
	"golang.org/x/exp/slog"
)

const (
	lightUpdateGracePeriod = 2 * time.Second
)

func (t *Timelight) handleEvent(event hue.Event) {
	t.log.Debug("handling event",
		slog.Any("event", event),
		slog.String("last_event_id", event.LastEventID),
		slog.Time("creation_time", event.CreationTime),
	)
	t.lastEventId = event.LastEventID

	switch event.Type {
	case "update":
		for _, resource := range event.Data {
			t.handleUpdate(resource, event.CreationTime)
		}

	default:
		t.log.Debug("unknown event type", slog.String("type", event.Type))
	}
}

func (t *Timelight) handleUpdate(res hue.Resource, eventTime time.Time) {
	switch r := res.(type) {
	case *hue.Scene:
		if r == nil {
			return
		}
		t.handleSceneUpdate(*r)

	case *hue.Light:
		if r == nil {
			return
		}
		t.handleLightUpdate(*r, eventTime)

	default:
		t.log.Debug("unkown resource type",
			slog.String("type", fmt.Sprintf("%T", res)),
			slog.String("rtype", string(res.Type())),
		)
	}
}

func (t *Timelight) handleSceneUpdate(scene hue.Scene) {
	tlScene, found := t.scenes[SceneID(scene.ID)]
	if !found {
		return // If it's not a timelight scene then we don't care about it.
		// TODO: If a non-timelight scene was recalled, DEACTIVATE lights that were
		// affected.
	}

	// TODO: Detect whether lights were added to/removed from scene?
	if scene.Status == nil {
		return // Does this mean scene was not recalled?
	}

	// Scene was (maybe?) recalled. Mark all lights as active.
	for _, light := range tlScene.Lights {
		light.SetActive()
		light.LastUpdated = time.Now()
		light.TargetState = tlScene.TargetState
	}

	t.log.Info("timelight scene recalled, marked lights as active",
		slog.String("scene_id", scene.ID),
		slog.Int("lights", len(tlScene.Lights)))
}

func (t *Timelight) handleLightUpdate(lightUpdate hue.Light, eventTime time.Time) {
	// TODO: Detect whether the light was updated by something other than timelight.
	// TODO: Do we need to handle grouped_light events seprately?

	t.log.Debug("checking for manual light change", slog.String("id", lightUpdate.ID))

	tlLight, found := t.lights[LightID(lightUpdate.ID)]
	if !found {
		return
	}
	if !tlLight.Active {
		return // Light is not active; this method won't change it.
	}

	if lightChanged(tlLight, lightUpdate, eventTime) {
		t.log.Info("changed detected",
			slog.String("id", string(tlLight.ID)),
			slog.Any("target", tlLight.TargetState),
			slog.Any("update_dimming", lightUpdate.Dimming),
			slog.Any("update_temp", lightUpdate.ColorTemperature),
		)
		tlLight.SetInactive()
	}
}

func lightChanged(light *Light, lightUpdate hue.Light, eventTime time.Time) bool {
	if lightUpdate.ID != string(light.ID) {
		return false
	}

	if lightUpdate.On != nil && !lightUpdate.On.On {
		return true // Light was turned off.
	}

	target := light.TargetState

	if light.HasBrightness && lightUpdate.Dimming != nil && target.HasBrightness {
		targetBrightness := target.Brightness
		updateBrightness := lightUpdate.Dimming.Brightness

		if !within(targetBrightness, updateBrightness, 2) {
			return true
		}
	}

	if light.HasColorTemperature && lightUpdate.ColorTemperature != nil && target.HasTempMirek {
		if !lightUpdate.ColorTemperature.MirekValid {
			return true
		}

		targetTemp := float64(light.TargetState.TempMirek)
		updateTemp := float64(lightUpdate.ColorTemperature.Mirek)

		if !within(targetTemp, updateTemp, 10) {
			return true
		}
	}

	return false
}

func within(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}
