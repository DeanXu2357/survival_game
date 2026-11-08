package system

import (
	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/system/ai"
)

var _ state.System = (*EnemyAISystem)(nil)

// EnemyAISystem walks all entities tagged with AIBehavior and dispatches
// each to the Behavior strategy registered under its Key. Per-tick state
// is owned by the strategy, not by this system — adding a new enemy means
// implementing Behavior and registering its key, with no changes here.
type EnemyAISystem struct {
	world       *state.World
	currentTick *ports.Tick
	registry    map[string]ai.Behavior
}

func NewEnemyAISystem(world *state.World, currentTick *ports.Tick, registry map[string]ai.Behavior) *EnemyAISystem {
	return &EnemyAISystem{
		world:       world,
		currentTick: currentTick,
		registry:    registry,
	}
}

func (s *EnemyAISystem) ReadMeta() state.Meta {
	return state.ComponentAIBehavior | state.ComponentPosition | state.ComponentDirection |
		state.ComponentHealth | state.ComponentInventory | state.ComponentCollider
}

func (s *EnemyAISystem) WriteMeta() state.Meta {
	return state.ComponentInput
}

func (s *EnemyAISystem) Update(dt float64) {
	ctx := ai.Context{World: s.world, CurrentTick: *s.currentTick, DT: dt}
	for id, ab := range s.world.AI.All() {
		behavior, ok := s.registry[ab.Key]
		if !ok {
			continue
		}
		behavior.Update(ctx, id)
	}
}

func (s *EnemyAISystem) Indicators() []state.ThreatIndicator {
	ctx := ai.Context{World: s.world, CurrentTick: *s.currentTick}
	out := make([]state.ThreatIndicator, 0)
	for id, ab := range s.world.AI.All() {
		behavior, ok := s.registry[ab.Key]
		if !ok {
			continue
		}
		if ind, visible := behavior.Indicator(ctx, id); visible {
			out = append(out, ind)
		}
	}
	return out
}
