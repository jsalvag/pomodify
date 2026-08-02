package session

import (
	"errors"
	"time"
)

var (
	ErrInvalidWorkDuration = errors.New("work duration must be positive")
	ErrInvalidRestDuration = errors.New("rest duration must be positive")
	ErrInvalidWorkBlocks   = errors.New("work blocks must be positive")
)

type PhaseKind string

const (
	WorkPhase PhaseKind = "work"
	RestPhase PhaseKind = "rest"
)

type Phase struct {
	Kind        PhaseKind
	Duration    time.Duration
	BlockNumber int
}

type Plan struct {
	Phases     []Phase
	WorkBlocks int
}

func BuildPlan(workDuration, restDuration time.Duration, workBlocks int) (Plan, error) {
	if workDuration <= 0 {
		return Plan{}, ErrInvalidWorkDuration
	}

	if restDuration <= 0 {
		return Plan{}, ErrInvalidRestDuration
	}

	if workBlocks <= 0 {
		return Plan{}, ErrInvalidWorkBlocks
	}

	phases := make([]Phase, 0, workBlocks*2-1)
	for blockNumber := 1; blockNumber <= workBlocks; blockNumber++ {
		phases = append(phases, Phase{
			Kind:        WorkPhase,
			Duration:    workDuration,
			BlockNumber: blockNumber,
		})

		if blockNumber == workBlocks {
			continue
		}

		phases = append(phases, Phase{
			Kind:        RestPhase,
			Duration:    restDuration,
			BlockNumber: blockNumber,
		})
	}

	return Plan{
		Phases:     phases,
		WorkBlocks: workBlocks,
	}, nil
}
