package turret

import (
	"survival/internal/engine/state"
	"survival/internal/engine/system/ai/goap"
)

var _ goap.ActionOnSymbolicWithRuntime = (*action)(nil)

// action is a symbolic GOAP action whose runtime step simply emits an Input. The
// turret's two actions (aim, fire) differ only in their symbols and whether they
// pull the trigger.
type action struct {
	name  string
	pre   goap.SymbolSet
	set   goap.SymbolSet
	clear goap.SymbolSet
	cost  float64
	fire  bool
}

func (a *action) Name() string                        { return a.name }
func (a *action) Preconditions() goap.SymbolSet       { return a.pre }
func (a *action) Effect() (set, clear goap.SymbolSet) { return a.set, a.clear }
func (a *action) StaticCost() float64                 { return a.cost }

func (a *action) Step(p goap.Perception, bb goap.BlackBoard, id state.EntityID) state.Input {
	return state.Input{
		PickupEntityID: state.NoPickup,
		DropSlotIndex:  state.NoDrop,
		Fire:           a.fire,
	}
}

func newActions() []goap.ActionOnSymbolic {
	var empty goap.SymbolSet
	aim := &action{
		name: "aim",
		pre:  empty.With(symHasTarget),
		set:  empty.With(symAimed),
		cost: 1,
	}
	fire := &action{
		name: "fire",
		pre:  empty.With(symHasTarget).With(symAimed).With(symWeaponReady),
		set:  empty.With(symTargetDown),
		cost: 1,
		fire: true,
	}
	return []goap.ActionOnSymbolic{aim, fire}
}
