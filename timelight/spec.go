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

func minuteOfDay(t time.Time) int {
	hour, min, _ := t.Clock()
	return 60*hour + min
}

type SmoothSpec struct {
	Brightness SmoothTransition
	TempMirek  SmoothTransition
}

func (s SmoothSpec) TargetLightState(now time.Time) TargetState {
	brightness := s.Brightness.targetValue(now)
	temp := s.TempMirek.targetValue(now)
	return DefaultTargetState.WithBrightness(brightness).WithColorTemp(int(temp))
}

type SmoothTransition struct {
	StartMinute      int // Time to begin applying start state.
	TransitionMinute int // Time to begin transition towards end state.
	EndMinute        int // Time to begin applying end state.

	Start float64
	End   float64
}

func (s SmoothTransition) progressPct(curMinute int) float64 {
	p := float64(curMinute-s.TransitionMinute) / float64(s.EndMinute-s.TransitionMinute)
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

func (s SmoothTransition) targetValue(now time.Time) float64 {
	curMinute := minuteOfDay(now)
	if curMinute < s.StartMinute || curMinute > s.EndMinute {
		return s.End
	}
	if curMinute < s.TransitionMinute {
		return s.Start
	}

	p := s.progressPct(curMinute)
	return transition(s.Start, s.End, p)
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
