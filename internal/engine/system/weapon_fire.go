package system

import (
	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/vector"
)

const (
	defaultProjectileHeight = 1.5
)

var _ state.System = (*WeaponFireSystem)(nil)

type WeaponFireSystem struct {
	world       *state.World
	currentTick *uint64
}

func NewWeaponFireSystem(world *state.World, currentTick *uint64) *WeaponFireSystem {
	return &WeaponFireSystem{world: world, currentTick: currentTick}
}

func (wf *WeaponFireSystem) ReadMeta() state.Meta {
	return state.ComponentInput | state.ComponentPosition | state.ComponentDirection | state.ComponentInventory
}

func (wf *WeaponFireSystem) WriteMeta() state.Meta {
	return state.ComponentProjectile | state.ComponentPosition | state.ComponentDirection |
		state.ComponentMeta | state.ComponentInventory
}

func (wf *WeaponFireSystem) Update(dt float64) {
	world := wf.world
	tick := *wf.currentTick

	for entityID, input := range world.Input.All() {
		if !input.Fire {
			continue
		}

		inv, invOk := world.Inventory.Get(entityID)
		if !invOk {
			continue
		}

		pos, posOk := world.Position.Get(entityID)
		dir, dirOk := world.Direction.Get(entityID)
		if !posOk || !dirOk {
			continue
		}

		weapon := wf.resolveActiveWeapon(inv)
		activeSlot := inv.Weapons[inv.CurrentWeaponIndex]

		// Fire rate limiting (LastFireTick == 0 means never fired, always allow)
		fireInterval := uint64(ports.TargetTickRate / weapon.FireRate)
		if activeSlot.LastFireTick > 0 && tick-activeSlot.LastFireTick < fireInterval {
			continue
		}

		wf.fireProjectile(world, entityID, pos, dir, weapon, tick)

		// Update per-weapon LastFireTick
		inv.Weapons[inv.CurrentWeaponIndex] = state.WeaponSlot{
			ItemDefID:    activeSlot.ItemDefID,
			LastFireTick: tick,
		}
		world.UpdatePlayer(entityID, state.UpdatePlayer{
			UpdateMeta: state.ComponentInventory,
			Inventory:  inv,
		})
	}
}

// resolveActiveWeapon looks up the WeaponSpec for the currently equipped weapon.
// Falls back to FistSpec if the active slot is empty or the ItemDef is missing.
func (wf *WeaponFireSystem) resolveActiveWeapon(inv state.Inventory) state.WeaponSpec {
	slot := inv.Weapons[inv.CurrentWeaponIndex]
	if slot.IsEmpty() {
		return state.FistSpec
	}

	itemDef, ok := wf.world.ItemDef.Get(slot.ItemDefID)
	if !ok {
		return state.FistSpec
	}

	// If the WeaponSpec is zero-valued, fall back to FistSpec
	if itemDef.WeaponSpec == (state.WeaponSpec{}) {
		return state.FistSpec
	}

	return itemDef.WeaponSpec
}

func (wf *WeaponFireSystem) fireProjectile(world *state.World, ownerID state.EntityID, pos state.Position, dir state.Direction, weapon state.WeaponSpec, tick uint64) {
	spawnPos := state.Position(vector.Vector2D(pos).Add(vector.Forward(float64(dir)).Scale(0.5)))

	speed := weapon.Speed
	expiredAt := tick + uint64(weapon.Range/speed*ports.TargetTickRate)

	world.CreateProjectileEntity(state.CreateProjectile{
		Position:  spawnPos,
		Direction: dir,
		ProjectileData: state.ProjectileData{
			Speed:     speed,
			Range:     weapon.Range,
			Damage:    weapon.Damage,
			OwnerID:   ownerID,
			Height:    defaultProjectileHeight,
			ExpiredAt: expiredAt,
		},
	})
}
