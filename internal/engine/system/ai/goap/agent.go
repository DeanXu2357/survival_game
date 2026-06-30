package goap

import (
	"survival/internal/engine/ports"
	"survival/internal/engine/state"
)

var _ Agent = (*BasicAgent)(nil)

const replanLimit uint8 = 3

type BasicAgentConfig struct {
	Planner   Planner
	Goal      Goal
	Actions   []ActionOnSymbolic
	Observer  Perception
	Projector Project
	Memory    BlackBoard
}

func NewBasicAgent(cfg BasicAgentConfig) *BasicAgent {
	return &BasicAgent{
		planner:   cfg.Planner,
		goal:      cfg.Goal,
		actions:   cfg.Actions,
		observer:  cfg.Observer,
		projector: cfg.Projector,
		memory:    cfg.Memory,
	}
}

type BasicAgent struct {
	planner   Planner
	goal      Goal
	actions   []ActionOnSymbolic
	observer  Perception
	projector Project
	memory    BlackBoard
	queue     []ActionOnSymbolicWithRuntime
}

func (a *BasicAgent) NextInput(world *state.World, currentTick ports.Tick, dt float64, id state.EntityID) (state.Input, bool) {
	a.observer.SetWorld(world)
	curState := a.projector(a.observer, id)

	if len(a.queue) == 0 {
		// replan
		if !a.doPlan(curState) {
			return state.Input{}, false
		}
	}

	var replanTimes uint8
	for len(a.queue) > 0 {
		act := a.queue[0]

		willSet, _ := act.Effect()
		if curState.Contains(willSet) {
			// pop
			a.queue = a.queue[1:]
			continue
		}

		// contains precondition ?
		if !curState.Contains(act.Preconditions()) {
			if replanTimes > replanLimit {
				return state.Input{}, false
			}

			if planned := a.doPlan(curState); !planned {
				// TODO: metric replan failed
				return state.Input{}, false
			}

			replanTimes++
			continue
		}

		return act.Step(a.observer, a.memory, id), true
	}
	return state.Input{}, false
}

func (a *BasicAgent) doPlan(curState SymbolSet) bool {
	actions, ok := a.planner.Plan(curState, a.goal, a.actions)
	if !ok {
		return false
	}

	a.queue = nil
	for _, act := range actions {
		rAct, isRuntime := act.(ActionOnSymbolicWithRuntime)
		if !isRuntime {
			// TODO: warns planner generating invalid actions
			// TODO: metric fails
			return false
		}
		a.queue = append(a.queue, rAct)
	}
	return true
}
