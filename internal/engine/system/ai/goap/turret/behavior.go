package turret

import (
	"math"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/system/ai"
	"survival/internal/engine/system/ai/goap"
)

const GOAPTurretKey = "goap_turret"

const (
	turretAimSeconds      = 1.0
	turretCooldownSeconds = 0.5
	turretConeHalfAngle   = math.Pi / 6 // 30° → 60° full cone
)

var (
	turretAimDuration = ports.TicksFromSeconds(turretAimSeconds)
	turretCooldown    = ports.TicksFromSeconds(turretCooldownSeconds)
)

var _ ai.Behavior = (*Behavior)(nil)

// bundle is the per-entity GOAP runtime. The AI registry shares one Behavior
// across all turrets, but each turret needs its own plan queue (BasicAgent),
// sensing (perception) and memory (blackboard).
type bundle struct {
	agent *goap.BasicAgent
	perc  *perception
	bb    *blackboard
}

// Behavior drives stationary turrets via the GOAP framework. The planner, goal,
// actions and projector are stateless and shared; only the bundle is per-entity.
type Behavior struct {
	agents    map[state.EntityID]*bundle
	planner   goap.Planner
	goal      goap.Goal
	actions   []goap.ActionOnSymbolic
	projector goap.Project
}

func NewBehavior() *Behavior {
	return &Behavior{
		agents:    make(map[state.EntityID]*bundle),
		planner:   newPlanner(),
		goal:      newGoal(),
		actions:   newActions(),
		projector: newProjector(),
	}
}

func (b *Behavior) Update(ctx ai.Context, id state.EntityID) {
	world := ctx.World

	if !world.Entity.IsAlive(id) {
		delete(b.agents, id)
		return
	}
	hp, hpOk := world.Health.Get(id)
	if !hpOk || int(hp) <= 0 {
		delete(b.agents, id)
		world.Input.Upsert(id, neutralInput())
		return
	}

	bd := b.bundleFor(id)
	input, ok := bd.agent.NextInput(world, ctx.CurrentTick, ctx.DT, id)
	if !ok {
		// Mandatory: when the agent yields no action (idle, cooldown, target
		// lost) we must still clear Fire. Otherwise a stale Input.Fire=true
		// latches and the weapon auto-refires at its own cooldown with no re-aim.
		input = neutralInput()
	}
	world.Input.Upsert(id, input)
}

func (b *Behavior) Indicator(ctx ai.Context, id state.EntityID) (state.ThreatIndicator, bool) {
	world := ctx.World
	if !world.Entity.IsAlive(id) {
		return state.ThreatIndicator{}, false
	}

	hp, hpOk := world.Health.Get(id)
	if !hpOk || int(hp) <= 0 {
		return state.ThreatIndicator{}, false
	}

	pos, posOk := world.Position.Get(id)
	dir, dirOk := world.Direction.Get(id)
	if !posOk || !dirOk {
		return state.ThreatIndicator{}, false
	}
	return state.ThreatIndicator{
		EntityID:  id,
		Shape:     state.IndicatorShapeCone,
		X:         pos.X,
		Y:         pos.Y,
		Direction: float64(dir),
		HalfAngle: turretConeHalfAngle,
		Range:     state.TurretSpec.Range,
	}, true
}

func (b *Behavior) bundleFor(id state.EntityID) *bundle {
	if bd, ok := b.agents[id]; ok {
		return bd
	}
	perc := newPerception()
	bb := newBlackboard()
	bd := &bundle{
		agent: goap.NewBasicAgent(goap.BasicAgentConfig{
			Planner:   b.planner,
			Goal:      b.goal,
			Actions:   b.actions,
			Observer:  perc,
			Projector: b.projector,
			Memory:    bb,
		}),
		perc: perc,
		bb:   bb,
	}
	b.agents[id] = bd
	return bd
}

func neutralInput() state.Input {
	return state.Input{
		PickupEntityID: state.NoPickup,
		DropSlotIndex:  state.NoDrop,
	}
}
