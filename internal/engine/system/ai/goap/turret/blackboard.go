package turret

import (
	"survival/internal/engine/ports"
	"survival/internal/engine/system/ai/goap"
	"survival/internal/engine/vector"
)

var _ goap.BlackBoard = (*blackboard)(nil)

// blackboard is the turret's per-entity memory: the aim state machine. The
// turret has no destination, so the goap.BlackBoard destination methods are
// no-ops.
type blackboard struct {
	aiming   bool
	aimStart ports.Tick
}

func newBlackboard() *blackboard { return &blackboard{} }

func (b *blackboard) SetDestination(v vector.Vector2D) bool { return false }

func (b *blackboard) Destination() (vector.Vector2D, bool) { return vector.Vector2D{}, false }

// UpdateAim advances the aim timer. Aim resets whenever the turret has no target
// or the weapon is on cooldown; otherwise it latches the tick aiming began.
func (b *blackboard) UpdateAim(hasTarget, weaponReady bool, tick ports.Tick) {
	switch {
	case !hasTarget || !weaponReady:
		b.aiming = false
	case !b.aiming:
		b.aiming = true
		b.aimStart = tick
	}
}

func (b *blackboard) Aimed(tick, aimDuration ports.Tick) bool {
	return b.aiming && tick-b.aimStart >= aimDuration
}
