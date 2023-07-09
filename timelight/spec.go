package timelight

import (
	"math"
	"time"
)

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
	if b, ok := s.brightness(p); ok {
		target = target.WithBrightness(b)
	}
	if t, ok := s.temp(p); ok {
		target = target.WithColorTemp(t)
	}

	return target
}

func (s LinearSpec) brightness(p float64) (float64, bool) {
	if !s.Start.HasBrightness && !s.End.HasBrightness {
		return 0, false
	}
	if !s.End.HasBrightness {
		return s.Start.Brightness, true
	}
	if !s.Start.HasBrightness {
		return s.End.Brightness, true
	}

	return s.Start.Brightness + p*(s.End.Brightness-s.Start.Brightness), true
}

func (s LinearSpec) temp(p float64) (int, bool) {
	if !s.Start.HasTempMirek && !s.End.HasTempMirek {
		return 0, false
	}
	if !s.End.HasTempMirek {
		return s.Start.TempMirek, true
	}
	if !s.Start.HasTempMirek {
		return s.End.TempMirek, true
	}

	return int(float64(s.Start.TempMirek) + p*float64(s.End.TempMirek-s.Start.TempMirek)), true
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

type SmoothSpec struct {
	StartMinute      int // Time to begin applying start state.
	TransitionMinute int // Time to begin transition towards end state.
	EndMinute        int // Time to begin applying end state.

	Start TargetState
	End   TargetState
}

func (s SmoothSpec) progressPct(curMinute int) float64 {
	p := float64(curMinute-s.TransitionMinute) / float64(s.EndMinute-s.TransitionMinute)
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

func (s SmoothSpec) TargetLightState(now time.Time) TargetState {
	curMinute := minuteOfDay(now)
	if curMinute < s.StartMinute || curMinute > s.EndMinute {
		return s.End
	}
	if curMinute < s.TransitionMinute {
		return s.Start
	}

	p := s.progressPct(curMinute)

	var target TargetState
	if b, ok := s.brightness(p); ok {
		target = target.WithBrightness(b)
	}
	if t, ok := s.temp(p); ok {
		target = target.WithColorTemp(t)
	}

	return target
}

func (s SmoothSpec) brightness(p float64) (float64, bool) {
	if !s.Start.HasBrightness && !s.End.HasBrightness {
		return 0, false
	}
	if !s.End.HasBrightness {
		return s.Start.Brightness, true
	}
	if !s.Start.HasBrightness {
		return s.End.Brightness, true
	}

	return transition(s.Start.Brightness, s.End.Brightness, p), true
}

func (s SmoothSpec) temp(p float64) (int, bool) {
	if !s.Start.HasTempMirek && !s.End.HasTempMirek {
		return 0, false
	}
	if !s.End.HasTempMirek {
		return s.Start.TempMirek, true
	}
	if !s.Start.HasTempMirek {
		return s.End.TempMirek, true
	}

	return int(transition(float64(s.Start.TempMirek), float64(s.End.TempMirek), p)), true
}

// For fixed values of start and end, transitions "smoothly" from start to end
// as p progresses from 0 to 1.
func transition(start, end, p float64) float64 {
	if p < 0 {
		return start
	}
	if p > 1 {
		return end
	}

	return 0.5*(start-end)*(1+math.Cos(math.Pi*p)) + end
}
