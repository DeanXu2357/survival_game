package turret

import (
	"math"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/system/ai"
)

const (
	TurretAimSeconds      = 1.0
	TurretCooldownSeconds = 0.5
	TurretConeHalfAngle   = math.Pi / 6 // 30° → 60° full cone
)

const StationaryTurretKey = "stationary_turret"

type StationaryTurretBehavior struct {
	aimStart map[state.EntityID]ports.Tick
}

func NewStationaryTurretBehavior() *StationaryTurretBehavior {
	return &StationaryTurretBehavior{
		aimStart: make(map[state.EntityID]ports.Tick),
	}
}

func (b *StationaryTurretBehavior) Update(ctx ai.Context, id state.EntityID) {
	world := ctx.World
	tick := *ctx.CurrentTick
	aimDuration := ports.TicksFromSeconds(TurretAimSeconds)
	cooldownDuration := ports.TicksFromSeconds(TurretCooldownSeconds)

	if !world.Entity.IsAlive(id) {
		delete(b.aimStart, id)
		return
	}

	hp, hpOk := world.Health.Get(id)
	if !hpOk || int(hp) <= 0 {
		delete(b.aimStart, id)
		b.clearFire(world, id)
		return
	}

	pos, posOk := world.Position.Get(id)
	dir, dirOk := world.Direction.Get(id)
	inv, invOk := world.Inventory.Get(id)
	if !posOk || !dirOk || !invOk {
		delete(b.aimStart, id)
		b.clearFire(world, id)
		return
	}

	target := b.acquireTarget(world, id, pos, float64(dir))
	if target == 0 {
		delete(b.aimStart, id)
		b.clearFire(world, id)
		return
	}

	slot := inv.Weapons[inv.CurrentWeaponIndex]
	if slot.LastFireTick > 0 && tick-slot.LastFireTick < cooldownDuration {
		delete(b.aimStart, id)
		b.clearFire(world, id)
		return
	}

	start, aiming := b.aimStart[id]
	if !aiming {
		b.aimStart[id] = tick
		b.clearFire(world, id)
		return
	}

	if tick-start >= aimDuration {
		b.setFire(world, id)
		delete(b.aimStart, id)
	} else {
		b.clearFire(world, id)
	}
}

func (b *StationaryTurretBehavior) Indicator(ctx ai.Context, id state.EntityID) (state.ThreatIndicator, bool) {
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
		HalfAngle: TurretConeHalfAngle,
		Range:     state.TurretSpec.Range,
	}, true
}

func (b *StationaryTurretBehavior) acquireTarget(world *state.World, self state.EntityID, pos state.Position, dir float64) state.EntityID {
	rng := state.TurretSpec.Range
	fx := math.Sin(dir)
	fy := -math.Cos(dir)
	cosLimit := math.Cos(TurretConeHalfAngle)

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
		targetPos, posOk := world.Position.Get(entityID)
		if !posOk {
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

func (b *StationaryTurretBehavior) setFire(world *state.World, id state.EntityID) {
	world.Input.Upsert(id, state.Input{
		PickupEntityID: state.NoPickup,
		DropSlotIndex:  state.NoDrop,
		Fire:           true,
	})
}

func (b *StationaryTurretBehavior) clearFire(world *state.World, id state.EntityID) {
	world.Input.Upsert(id, state.Input{
		PickupEntityID: state.NoPickup,
		DropSlotIndex:  state.NoDrop,
	})
}
