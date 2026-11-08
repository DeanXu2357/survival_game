package goap

import (
	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/vector"
)

type ActionOnSymbolic interface {
	Name() string
	Preconditions() SymbolSet
	Effect() (set SymbolSet, clear SymbolSet)
	StaticCost() float64
}

type ActionRuntime interface {
	Step(perception Perception, blackboard BlackBoard, id state.EntityID) state.Input
}

type ActionOnSymbolicWithRuntime interface {
	ActionOnSymbolic
	ActionRuntime
}

type Goal interface {
	Require() SymbolSet
}

type Planner interface {
	Plan(start SymbolSet, goal Goal, actions []ActionOnSymbolic) (plan []ActionOnSymbolic, ok bool)
}

// Agent is the runtime boundary: each tick it decides the input for an entity.
type Agent interface {
	NextInput(world *state.World, currentTick ports.Tick, dt float64, id state.EntityID) (state.Input, bool)
}

type Project func(perception Perception, blackboard BlackBoard, id state.EntityID) SymbolSet

// Frame carries the per-tick runtime context. It enters a behavior through
// Perception.Observe once per tick; projection and action runtime read it back
// via Perception.Frame, so the perception is the single source of truth for
// "now" (Tick/DT) and the current world.
type Frame struct {
	World *state.World
	Tick  ports.Tick
	DT    float64
}

type Perception interface {
	Observe(frame Frame)
	Frame() Frame
	Position(id state.EntityID) (vector.Vector2D, bool)
}

type BlackBoard interface {
	SetDestination(v vector.Vector2D) bool
	Destination() (vector.Vector2D, bool)
}
