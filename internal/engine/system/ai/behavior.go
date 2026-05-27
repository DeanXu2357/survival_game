package ai

import (
	"survival/internal/engine/ports"
	"survival/internal/engine/state"
)

type Context struct {
	World       *state.World
	CurrentTick *ports.Tick
	DT          float64
}

type Behavior interface {
	Update(ctx Context, id state.EntityID)
	Indicator(ctx Context, id state.EntityID) (state.ThreatIndicator, bool)
}
