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

		// Block firing while reloading.
		// Note: on the tick reload completes, InventorySystem clears ReloadStartTick
		// via command buffer, which isn't applied until after all systems run.
		// This means firing is delayed by one extra tick (~16ms), which is
		// imperceptible. Accepting this avoids coupling WeaponFireSystem to
		// reload duration logic.
		if activeSlot.ReloadStartTick > 0 {
			continue
		}

		// Fire rate limiting (LastFireTick == 0 means never fired, always allow)
		fireInterval := uint64(ports.TargetTickRate / weapon.FireRate)
		if activeSlot.LastFireTick > 0 && tick-activeSlot.LastFireTick < fireInterval {
			continue
		}

		// Check ammo for ranged weapons
		if weapon.AmmoCategory != state.AmmoCategoryNone {
			if activeSlot.LoadedMagAmmo <= 0 {
				continue // no ammo
			}
		}

		wf.fireProjectile(world, entityID, pos, dir, weapon, tick)

		// Update per-weapon LastFireTick and deduct ammo
		activeSlot.LastFireTick = tick
		if weapon.AmmoCategory != state.AmmoCategoryNone {
			activeSlot.LoadedMagAmmo--
		}
		inv.Weapons[inv.CurrentWeaponIndex] = activeSlot
		world.UpdatePlayer(entityID, state.UpdatePlayer{
			UpdateMeta: state.ComponentInventory,
			Inventory:  inv,
		})
	}
}

// resolveActiveWeapon looks up the WeaponConfig for the currently equipped weapon.
// Falls back to FistSpec if the active slot is empty or the ItemConfig is missing.
func (wf *WeaponFireSystem) resolveActiveWeapon(inv state.Inventory) state.WeaponConfig {
	slot := inv.Weapons[inv.CurrentWeaponIndex]
	if slot.IsEmpty() {
		return state.FistSpec
	}

	itemDef, ok := wf.world.ItemConfig.Get(slot.ItemID)
	if !ok {
		return state.FistSpec
	}

	// If the WeaponConfig is zero-valued, fall back to FistSpec
	if itemDef.WeaponConfig == (state.WeaponConfig{}) {
		return state.FistSpec
	}

	return itemDef.WeaponConfig
}

func (wf *WeaponFireSystem) fireProjectile(world *state.World, ownerID state.EntityID, pos state.Position, dir state.Direction, weapon state.WeaponConfig, tick uint64) {
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
