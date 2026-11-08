package turret

import (
	"testing"

	"survival/internal/engine/system/ai/goap"
)

func planNames(plan []goap.ActionOnSymbolic) []string {
	names := make([]string, len(plan))
	for i, a := range plan {
		names[i] = a.Name()
	}
	return names
}

func TestPlanner_AimedAndReadyPlansFireOnly(t *testing.T) {
	p := newPlanner()
	var start goap.SymbolSet
	start = start.With(symHasTarget).With(symAimed).With(symWeaponReady)

	plan, ok := p.Plan(start, newGoal(), newActions())
	if !ok {
		t.Fatalf("expected a plan")
	}
	if got := planNames(plan); len(got) != 1 || got[0] != "fire" {
		t.Fatalf("expected [fire], got %v", got)
	}
}

func TestPlanner_NotAimedPlansAimThenFire(t *testing.T) {
	p := newPlanner()
	var start goap.SymbolSet
	start = start.With(symHasTarget).With(symWeaponReady)

	plan, ok := p.Plan(start, newGoal(), newActions())
	if !ok {
		t.Fatalf("expected a plan")
	}
	got := planNames(plan)
	if len(got) != 2 || got[0] != "aim" || got[1] != "fire" {
		t.Fatalf("expected [aim fire], got %v", got)
	}
}

func TestPlanner_OnCooldownHasNoPlan(t *testing.T) {
	p := newPlanner()
	var start goap.SymbolSet
	start = start.With(symHasTarget) // no weapon-ready, nothing produces it

	if _, ok := p.Plan(start, newGoal(), newActions()); ok {
		t.Fatalf("expected no plan while on cooldown")
	}
}

func TestPlanner_NoTargetGoalAlreadyMet(t *testing.T) {
	p := newPlanner()
	var start goap.SymbolSet
	start = start.With(symTargetDown)

	plan, ok := p.Plan(start, newGoal(), newActions())
	if !ok {
		t.Fatalf("expected goal already satisfied")
	}
	if len(plan) != 0 {
		t.Fatalf("expected empty plan, got %v", planNames(plan))
	}
}
