package session

import (
	"testing"
	"time"
)

func TestBuildPlanCreatesAlternatingPhases(t *testing.T) {
	plan, err := BuildPlan(45*time.Minute, 15*time.Minute, 3)
	if err != nil {
		t.Fatalf("expected plan to build, got error: %v", err)
	}

	if len(plan.Phases) != 5 {
		t.Fatalf("expected 5 phases, got %d", len(plan.Phases))
	}

	if plan.Phases[0].Kind != WorkPhase || plan.Phases[0].BlockNumber != 1 {
		t.Fatalf("unexpected first phase: %+v", plan.Phases[0])
	}

	if plan.Phases[1].Kind != RestPhase || plan.Phases[1].BlockNumber != 1 {
		t.Fatalf("unexpected second phase: %+v", plan.Phases[1])
	}

	if plan.Phases[4].Kind != WorkPhase || plan.Phases[4].BlockNumber != 3 {
		t.Fatalf("unexpected final phase: %+v", plan.Phases[4])
	}
}

func TestBuildPlanRejectsInvalidInputs(t *testing.T) {
	if _, err := BuildPlan(0, 5*time.Minute, 1); err != ErrInvalidWorkDuration {
		t.Fatalf("expected invalid work duration error, got %v", err)
	}

	if _, err := BuildPlan(25*time.Minute, 0, 1); err != ErrInvalidRestDuration {
		t.Fatalf("expected invalid rest duration error, got %v", err)
	}

	if _, err := BuildPlan(25*time.Minute, 5*time.Minute, 0); err != ErrInvalidWorkBlocks {
		t.Fatalf("expected invalid work blocks error, got %v", err)
	}
}
