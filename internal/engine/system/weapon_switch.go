package system

import (
	"survival/internal/engine/state"
)

const (
	weaponSwitchCooldownTicks = 15 // ~0.25s at 60 FPS
)

var _ state.System = (*WeaponSwitchSystem)(nil)

type WeaponSwitchSystem struct {
	world       *state.World
	currentTick *uint64
}

func NewWeaponSwitchSystem(world *state.World, currentTick *uint64) *WeaponSwitchSystem {
	return &WeaponSwitchSystem{world: world, currentTick: currentTick}
}

func (ws *WeaponSwitchSystem) ReadMeta() state.Meta {
	return state.ComponentInput | state.ComponentWeaponState
}

func (ws *WeaponSwitchSystem) WriteMeta() state.Meta {
	return state.ComponentWeaponState
}

func (ws *WeaponSwitchSystem) Update(dt float64) {
	world := ws.world
	tick := *ws.currentTick

	for entityID, input := range world.Input.All() {
		if !input.SwitchWeapon {
			continue
		}

		ws, wsOk := world.WeaponState.Get(entityID)
		if !wsOk {
			continue
		}

		// Enforce cooldown
		if tick-ws.LastSwitchTick < weaponSwitchCooldownTicks {
			continue
		}

		// Cycle weapon index: 0 -> 1 -> 2 -> 0
		ws.CurrentWeaponIndex = (ws.CurrentWeaponIndex + 1) % len(ws.Weapons)
		ws.LastSwitchTick = tick

		world.UpdatePlayer(entityID, state.UpdatePlayer{
			UpdateMeta:  state.ComponentWeaponState,
			WeaponState: ws,
		})
	}
}
