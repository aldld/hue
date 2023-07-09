package timelight

import (
	"strings"
	"time"

	"github.com/aldld/hue/hue"
	"golang.org/x/exp/slog"
)

type SceneID string

type Scene struct {
	t *Timelight

	ID          SceneID
	TargetState TargetState
	LastUpdated time.Time
	Hue         hue.Scene
	Lights      map[LightID]*Light
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
			t: t,

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
		if oldAction.Target.Type != hue.RTypeLight {
			continue
		}
		lid := LightID(oldAction.Target.ID)
		light, ok := s.Lights[lid]
		if !ok {
			continue
		}

		newActions[i].Target = oldAction.Target
		newActions[i].Action.On = &hue.LightOn{On: true}

		if light.HasBrightness && lightState.HasBrightness {
			newActions[i].Action.Dimming = &hue.DimmingAction{
				Brightness: lightState.Brightness,
			}
		}
		if light.HasColorTemperature && lightState.HasTempMirek {
			newActions[i].Action.ColorTemperature = &hue.ColorTemperatureAction{
				Mirek: lightState.TempMirek,
			}
		}
	}

	err := s.t.hue.UpdateScene(s.Hue.ID, hue.SceneUpdate{Actions: &newActions})
	if err != nil {
		return err
	}

	s.TargetState = lightState
	s.LastUpdated = time.Now()

	return nil
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
