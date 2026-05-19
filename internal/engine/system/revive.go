package system

import (
	"survival/internal/engine/ports"
	"survival/internal/engine/state"
)

var _ state.System = (*ReviveSystem)(nil)

// ReviveSystem drives death/respawn for any entity carrying a Revive
// component. On death it stamps DeadAt; after RespawnDelayTicks have
// elapsed it refills Health and clears DeadAt. Hit detection skips
// entities with Health <= 0, so dead entities naturally stop absorbing
// damage during the respawn window without component surgery.
type ReviveSystem struct {
	world       *state.World
	currentTick *ports.Tick
}

func NewReviveSystem(world *state.World, currentTick *ports.Tick) *ReviveSystem {
	return &ReviveSystem{world: world, currentTick: currentTick}
}

func (rs *ReviveSystem) ReadMeta() state.Meta {
	return state.ComponentRevive | state.ComponentHealth
}

func (rs *ReviveSystem) WriteMeta() state.Meta {
	return state.ComponentHealth
}

func (rs *ReviveSystem) Update(dt float64) {
	world := rs.world
	tick := *rs.currentTick

	for entityID, revive := range world.Revive.All() {
		if revive.DeadAt == 0 {
			hp, ok := world.Health.Get(entityID)
			if !ok || int(hp) > 0 {
				continue
			}

			revive.DeadAt = tick
			world.Revive.Set(entityID, revive)
			continue
		}

		if tick-revive.DeadAt < revive.RespawnDelayTicks {
			continue
		}

		revive.DeadAt = 0
		world.Revive.Set(entityID, revive)
		world.Health.Set(entityID, revive.SpawnHealth)
	}
}
