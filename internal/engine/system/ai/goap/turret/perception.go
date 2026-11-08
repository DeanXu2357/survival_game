package turret

import (
	"math"

	"survival/internal/engine/state"
	"survival/internal/engine/system/ai/goap"
	"survival/internal/engine/vector"
)

var _ goap.Perception = (*perception)(nil)

// perception is pure sensing: it holds the per-tick frame handed in via Observe
// and answers world queries (target acquisition, weapon readiness) from it.
type perception struct {
	frame goap.Frame
}

func newPerception() *perception { return &perception{} }

func (p *perception) Observe(frame goap.Frame) { p.frame = frame }

func (p *perception) Frame() goap.Frame { return p.frame }

func (p *perception) Position(id state.EntityID) (vector.Vector2D, bool) {
	if p.frame.World == nil {
		return vector.Vector2D{}, false
	}
	pos, ok := p.frame.World.Position.Get(id)
	if !ok {
		return vector.Vector2D{}, false
	}
	return vector.Vector2D{X: pos.X, Y: pos.Y}, true
}

// acquire returns the first valid target inside the turret's frontal cone and
// range, reproducing the legacy turret's targeting rules: a non-self, non-AI
// circular entity with positive health.
func (p *perception) acquire(self state.EntityID) state.EntityID {
	world := p.frame.World
	if world == nil {
		return 0
	}
	pos, posOk := world.Position.Get(self)
	dir, dirOk := world.Direction.Get(self)
	if !posOk || !dirOk {
		return 0
	}

	rng := state.TurretSpec.Range
	fx := math.Sin(float64(dir))
	fy := -math.Cos(float64(dir))
	cosLimit := math.Cos(turretConeHalfAngle)

	for entityID, col := range world.Collider.All() {
		if entityID == self {
			continue
		}
		if col.ShapeType != state.ColliderCircle {
			continue
		}
		hp, hpOk := world.Health.Get(entityID)
		if !hpOk || int(hp) <= 0 {
			continue
		}
		if _, isAI := world.AI.Get(entityID); isAI {
			continue
		}
		targetPos, ok := world.Position.Get(entityID)
		if !ok {
			continue
		}
		dx := targetPos.X - pos.X
		dy := targetPos.Y - pos.Y
		dist2 := dx*dx + dy*dy
		if dist2 > rng*rng {
			continue
		}
		dist := math.Sqrt(dist2)
		if dist < 1e-6 {
			return entityID
		}
		cosAngle := (fx*dx + fy*dy) / dist
		if cosAngle >= cosLimit {
			return entityID
		}
	}
	return 0
}

// weaponReady reports whether the turret's equipped weapon is off cooldown,
// using the real per-weapon LastFireTick written by the weapon fire system.
func (p *perception) weaponReady(self state.EntityID) bool {
	world := p.frame.World
	if world == nil {
		return false
	}
	inv, ok := world.Inventory.Get(self)
	if !ok {
		return false
	}
	slot := inv.Weapons[inv.CurrentWeaponIndex]
	if slot.LastFireTick == 0 {
		return true
	}
	return p.frame.Tick-slot.LastFireTick >= turretCooldown
}
