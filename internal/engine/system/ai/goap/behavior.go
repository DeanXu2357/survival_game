package goap

import (
	"survival/internal/engine/state"
	"survival/internal/engine/system/ai"
)

var _ ai.Behavior = (*Behavior)(nil)

func NewGOAPBehavior(agent Agent) *Behavior {
	return &Behavior{agent: agent}
}

type Behavior struct {
	agent Agent
}

func (b *Behavior) Update(ctx ai.Context, id state.EntityID) {
	input, ok := b.agent.NextInput(ctx.World, ctx.CurrentTick, ctx.DT, id)
	if ok {
		ctx.World.Input.Upsert(id, input)
	}
}

func (b *Behavior) Indicator(ctx ai.Context, id state.EntityID) (state.ThreatIndicator, bool) {
	return state.ThreatIndicator{}, false
}
