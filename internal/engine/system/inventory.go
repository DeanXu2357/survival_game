package system

import (
	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/vector"
)

const (
	weaponSwitchCooldownTicks = 15 // ~0.25s at 60 FPS
)

var _ state.System = (*InventorySystem)(nil)

type InventorySystem struct {
	world       *state.World
	currentTick *uint64
}

func NewInventorySystem(world *state.World, currentTick *uint64) *InventorySystem {
	return &InventorySystem{world: world, currentTick: currentTick}
}

func (is *InventorySystem) ReadMeta() state.Meta {
	return state.ComponentInput | state.ComponentPosition | state.ComponentInventory | state.ComponentDirection
}

func (is *InventorySystem) WriteMeta() state.Meta {
	return state.ComponentInventory | state.ComponentGroundItem | state.ComponentPosition | state.ComponentMeta
}

func (is *InventorySystem) Update(dt float64) {
	world := is.world

	for entityID, input := range world.Input.All() {
		inv, invOk := world.Inventory.Get(entityID)
		if !invOk {
			continue
		}

		updated := false

		if input.PickupEntityID != state.NoPickup {
			pos, posOk := world.Position.Get(entityID)
			if posOk {
				is.handlePickup(entityID, pos, &inv, state.EntityID(input.PickupEntityID))
				updated = true
			}
		}
		if input.DropSlotIndex != state.NoDrop {
			pos, posOk := world.Position.Get(entityID)
			if posOk {
				is.handleDrop(entityID, pos, &inv, input.DropSlotIndex)
				updated = true
			}
		}
		if input.SwitchWeapon {
			is.handleWeaponSwitch(entityID, &inv)
			updated = true
		}
		if is.handleReload(entityID, &inv, input) {
			updated = true
		}

		if updated {
			world.UpdatePlayer(entityID, state.UpdatePlayer{
				UpdateMeta: state.ComponentInventory,
				Inventory:  inv,
			})
		}
	}
}

func (is *InventorySystem) handlePickup(playerID state.EntityID, playerPos state.Position, inv *state.Inventory, targetEntityID state.EntityID) {
	if !is.world.Entity.IsAlive(targetEntityID) {
		return
	}

	gi, ok := is.world.GroundItem.Get(targetEntityID)
	if !ok {
		return
	}

	itemPos, posOk := is.world.Position.Get(targetEntityID)
	if !posOk {
		return
	}

	dist := vector.Vector2D(playerPos).DistanceTo(vector.Vector2D(itemPos))
	if dist > ports.PickupRange {
		return
	}

	// Look up the item definition to check if it's a weapon
	itemDef, defOk := is.world.ItemConfig.Get(gi.ItemDefID)
	if !defOk {
		return
	}

	if itemDef.Type == state.ItemTypeWeapon {
		// Route weapon to the correct weapon slot based on WeaponConfig.Type
		slotIdx := weaponSlotIndex(itemDef.WeaponConfig.Type)
		if slotIdx < 0 {
			return
		}
		if !inv.Weapons[slotIdx].IsEmpty() {
			return // slot already occupied
		}
		inv.Weapons[slotIdx] = state.WeaponSlot{ItemID: gi.ItemDefID}
	} else {
		// Non-weapon item goes to item slots
		slotIdx := findFreeItemSlot(*inv)
		if slotIdx < 0 {
			return
		}
		inv.Items[slotIdx] = state.ItemSlot{
			ItemDefID: gi.ItemDefID,
			Quantity:  gi.Quantity,
			Ammo:      gi.Ammo,
		}
	}

	is.world.QueueDestroyEntity(targetEntityID)
}

// weaponSlotIndex returns the weapon slot index for a given weapon type.
// Returns -1 for Fist (cannot be picked up) or unknown types.
func weaponSlotIndex(wt state.WeaponType) int {
	switch wt {
	case state.WeaponTypeKnife:
		return 1
	case state.WeaponTypeGun:
		return 2
	default:
		return -1
	}
}

func (is *InventorySystem) handleDrop(playerID state.EntityID, playerPos state.Position, inv *state.Inventory, slotIndex int) {
	if slotIndex < 0 {
		return
	}

	dir, dirOk := is.world.Direction.Get(playerID)
	if !dirOk {
		return
	}

	var itemDefID state.EntityID
	var quantity int
	var ammo int

	if slotIndex < 3 {
		// Weapon slot drop
		if slotIndex == 0 {
			return // cannot drop Fist
		}
		ws := inv.Weapons[slotIndex]
		if ws.IsEmpty() {
			return
		}
		itemDefID = ws.ItemID
		quantity = 1
		inv.Weapons[slotIndex] = state.WeaponSlot{}
	} else {
		// Item slot drop (index 3+ maps to Items[index-3])
		itemIdx := slotIndex - 3
		if itemIdx >= ports.ItemSlotCount {
			return
		}
		slot := inv.Items[itemIdx]
		if slot.IsEmpty() {
			return
		}
		itemDefID = slot.ItemDefID
		quantity = slot.Quantity
		ammo = slot.Ammo
		inv.Items[itemIdx] = state.ItemSlot{}
	}

	fwd := vector.Forward(float64(dir)).Scale(1.5)
	dropPos := vector.Vector2D(playerPos).Add(fwd)

	is.world.CreateGroundItemEntity(state.CreateGroundItem{
		Position:  state.Position(dropPos),
		ItemDefID: itemDefID,
		Quantity:  quantity,
		Ammo:      ammo,
	})
}

func (is *InventorySystem) handleWeaponSwitch(entityID state.EntityID, inv *state.Inventory) {
	tick := *is.currentTick

	if tick-inv.LastSwitchTick < weaponSwitchCooldownTicks {
		return
	}

	// Cancel any in-progress reload on current weapon
	activeSlot := &inv.Weapons[inv.CurrentWeaponIndex]
	if activeSlot.ReloadStartTick > 0 {
		activeSlot.ReloadStartTick = 0
		activeSlot.ReloadType = state.ReloadTypeNone
	}

	occupied := inv.OccupiedWeaponSlots()
	if len(occupied) <= 1 {
		return // nothing to switch to
	}

	// Find current position in occupied list and advance to next
	currentIdx := 0
	for i, slot := range occupied {
		if slot == inv.CurrentWeaponIndex {
			currentIdx = i
			break
		}
	}
	nextIdx := (currentIdx + 1) % len(occupied)
	inv.CurrentWeaponIndex = occupied[nextIdx]
	inv.LastSwitchTick = tick
}

func (is *InventorySystem) handleReload(entityID state.EntityID, inv *state.Inventory, input state.Input) bool {
	tick := *is.currentTick
	activeSlot := &inv.Weapons[inv.CurrentWeaponIndex]
	weaponConfig := is.resolveWeaponConfig(*inv)

	// Complete in-progress reload
	if activeSlot.ReloadStartTick > 0 {
		duration := reloadDuration(activeSlot.ReloadType)
		if tick-activeSlot.ReloadStartTick >= duration {
			is.completeReload(entityID, inv, weaponConfig)
			return true
		}
		return false
	}

	// Start new reload
	if !input.Reload && !input.FastReload {
		return false
	}
	if weaponConfig.AmmoCategory == state.AmmoCategoryNone {
		return false // melee cannot reload
	}
	if activeSlot.MagCapacity > 0 && activeSlot.LoadedMagAmmo >= activeSlot.MagCapacity {
		return false // magazine already full
	}
	magIdx := findFullestCompatibleMag(*inv, weaponConfig.AmmoCategory, is.world)
	if magIdx < 0 {
		return false // no spare magazine
	}

	reloadType := state.ReloadTypeNormal
	if input.FastReload {
		reloadType = state.ReloadTypeFast
	}

	activeSlot.ReloadStartTick = tick
	activeSlot.ReloadType = reloadType
	return true
}

func (is *InventorySystem) completeReload(entityID state.EntityID, inv *state.Inventory, weaponConfig state.WeaponConfig) {
	world := is.world
	activeSlot := &inv.Weapons[inv.CurrentWeaponIndex]

	// Handle old magazine
	if activeSlot.LoadedMagID != 0 {
		if activeSlot.ReloadType == state.ReloadTypeFast {
			// Fast reload: drop magazine on ground (with remaining ammo)
			if activeSlot.LoadedMagAmmo > 0 {
				pos, posOk := world.Position.Get(entityID)
				dir, dirOk := world.Direction.Get(entityID)
				if posOk && dirOk {
					fwd := vector.Forward(float64(dir)).Scale(1.5)
					dropPos := vector.Vector2D(pos).Add(fwd)
					world.CreateGroundItemEntity(state.CreateGroundItem{
						Position:  state.Position(dropPos),
						ItemDefID: activeSlot.LoadedMagID,
						Quantity:  1,
						Ammo:      activeSlot.LoadedMagAmmo,
					})
				}
			}
			// Empty mag is discarded
		} else {
			// Normal reload: put magazine back in inventory
			freeIdx := findFreeItemSlot(*inv)
			if freeIdx >= 0 {
				inv.Items[freeIdx] = state.ItemSlot{
					ItemDefID: activeSlot.LoadedMagID,
					Quantity:  1,
					Ammo:      activeSlot.LoadedMagAmmo,
				}
			}
		}
	}

	// Load new magazine from inventory (fullest first)
	magIdx := findFullestCompatibleMag(*inv, weaponConfig.AmmoCategory, world)
	if magIdx >= 0 {
		magSlot := inv.Items[magIdx]
		magDef, ok := world.ItemConfig.Get(magSlot.ItemDefID)
		if ok {
			activeSlot.LoadedMagID = magSlot.ItemDefID
			activeSlot.LoadedMagAmmo = magSlot.Ammo
			activeSlot.MagCapacity = magDef.MagCapacity
			inv.Items[magIdx] = state.ItemSlot{} // remove from inventory
		}
	} else {
		// No magazine available
		activeSlot.LoadedMagID = 0
		activeSlot.LoadedMagAmmo = 0
		activeSlot.MagCapacity = 0
	}

	// Clear reload state
	activeSlot.ReloadStartTick = 0
	activeSlot.ReloadType = state.ReloadTypeNone
}

func (is *InventorySystem) resolveWeaponConfig(inv state.Inventory) state.WeaponConfig {
	slot := inv.Weapons[inv.CurrentWeaponIndex]
	if slot.IsEmpty() {
		return state.FistSpec
	}

	itemDef, ok := is.world.ItemConfig.Get(slot.ItemID)
	if !ok {
		return state.FistSpec
	}

	if itemDef.WeaponConfig == (state.WeaponConfig{}) {
		return state.FistSpec
	}

	return itemDef.WeaponConfig
}

// findFullestCompatibleMag returns the index of the fullest compatible magazine in the inventory.
// Returns -1 if no compatible magazine is found.
func findFullestCompatibleMag(inv state.Inventory, ammoCategory state.AmmoCategory, world *state.World) int {
	bestIdx := -1
	bestAmmo := -1
	for i, slot := range inv.Items {
		if slot.IsEmpty() {
			continue
		}
		itemDef, ok := world.ItemConfig.Get(slot.ItemDefID)
		if !ok {
			continue
		}
		if itemDef.Type != state.ItemTypeMagazine {
			continue
		}
		if itemDef.AmmoCategory != ammoCategory {
			continue
		}
		if slot.Ammo > bestAmmo {
			bestAmmo = slot.Ammo
			bestIdx = i
		}
	}
	return bestIdx
}

// reloadDuration returns the number of ticks for a given reload type.
func reloadDuration(rt state.ReloadType) uint64 {
	switch rt {
	case state.ReloadTypeNormal:
		return ports.NormalReloadTicks
	case state.ReloadTypeFast:
		return ports.FastReloadTicks
	default:
		return 0
	}
}

func findFreeItemSlot(inv state.Inventory) int {
	limit := inv.MaxItemCapacity
	if limit <= 0 || limit > ports.ItemSlotCount {
		limit = ports.ItemSlotCount
	}
	for i := 0; i < limit; i++ {
		if inv.Items[i].IsEmpty() {
			return i
		}
	}
	return -1
}
