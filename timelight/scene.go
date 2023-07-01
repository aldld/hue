package timelight

import (
	"strings"

	"github.com/aldld/hue/hue"
	"golang.org/x/exp/slog"
)

type SceneID string

type Scene struct {
	h *hue.Client

	ID     SceneID
	Hue    hue.Scene
	Lights map[LightID]*Light
}

func (t *Timelight) initScenes() error {
	t.scenes = make(map[SceneID]*Scene) // Reset to empty map.

	scenes, err := t.hue.GetScenes()
	if err != nil {
		return err
	}

	for _, s := range scenes {
		if !isTimelightScene(s) {
			continue
		}

		scene := &Scene{
			h: t.hue,

			ID:     SceneID(s.ID),
			Hue:    s,
			Lights: make(map[LightID]*Light),
		}

		for _, action := range s.Actions {
			if action.Target.Type != hue.RTypeLight {
				continue
			}
			lightID := LightID(action.Target.ID)
			light, ok := t.lights[lightID]
			if !ok {
				t.log.Warn("Light not found", slog.String("ID", string(lightID)))
				continue
			}
			scene.Lights[lightID] = light
			light.Active = true
		}
		t.log.Info("Initialized scene",
			slog.String("name", scene.Hue.Metadata.Name),
			slog.Int("lights", len(scene.Lights)),
		)
		t.scenes[scene.ID] = scene
	}

	t.log.Info("Initialized timelight scenes", slog.Int("count", len(t.scenes)))

	return nil
}

func isTimelightScene(scene hue.Scene) bool {
	return strings.Contains(strings.ToLower(scene.Metadata.Name), "timelight")
}

func (s *Scene) UpdateActions(lightState TargetState) error {
	newActions := make([]hue.SceneAction, len(s.Hue.Actions))
	for i, oldAction := range s.Hue.Actions {
		newActions[i].Target = oldAction.Target

		if lightState.HasBrightness {
			newActions[i].Action.Dimming = &hue.DimmingAction{
				Brightness: lightState.Brightness,
			}
		}
		if lightState.HasTempMirek {
			newActions[i].Action.ColorTemperature = &hue.ColorTemperatureAction{
				Mirek: lightState.TempMirek,
			}
		}
	}

	return s.h.UpdateScene(s.Hue.ID, hue.SceneUpdate{Actions: &newActions})
}

func (t *Timelight) updateScenes(target TargetState) {
	t.log.Info("updating scenes", slog.Any("target", target))

	successes := 0
	errs := 0

	for _, scene := range t.scenes {
		if err := scene.UpdateActions(target); err != nil {
			t.log.Error("error while updating scene",
				slog.String("id", string(scene.ID)),
				slog.Any("err", err),
			)
			errs += 1
		} else {
			successes += 1
		}
	}
	t.log.Info("finished updating scenes",
		slog.Int("successes", successes),
		slog.Int("errs", errs),
	)
}
