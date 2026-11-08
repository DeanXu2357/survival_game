package turret

import (
	"survival/internal/engine/state"
	"survival/internal/engine/system/ai/goap"
)

// newProjector returns the turret's symbol projection: a state-free function of
// (sensing, memory) for the entity. It advances the aim timer in the blackboard
// and reports the current symbolic world state.
func newProjector() goap.Project {
	return func(p goap.Perception, bb goap.BlackBoard, id state.EntityID) goap.SymbolSet {
		perc, ok := p.(*perception)
		if !ok {
			return 0
		}
		board, ok := bb.(*blackboard)
		if !ok {
			return 0
		}

		tick := perc.Frame().Tick
		hasTarget := perc.acquire(id) != 0
		weaponReady := perc.weaponReady(id)
		board.UpdateAim(hasTarget, weaponReady, tick)

		var s goap.SymbolSet
		if hasTarget {
			s = s.With(symHasTarget)
		} else {
			s = s.With(symTargetDown)
		}
		if board.Aimed(tick, turretAimDuration) {
			s = s.With(symAimed)
		}
		if weaponReady {
			s = s.With(symWeaponReady)
		}
		return s
	}
}
