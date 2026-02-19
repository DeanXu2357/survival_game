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
	return state.ComponentInput | state.ComponentPosition | state.ComponentInventory
}

func (is *InventorySystem) WriteMeta() state.Meta {
	return state.ComponentInventory | state.ComponentGroundItem | state.ComponentPosition | state.ComponentMeta
}

func (is *InventorySystem) Update(dt float64) {
	world := is.world

	for entityID, input := range world.Input.All() {
		if input.PickupEntityID == state.NoPickup && input.DropSlotIndex == state.NoDrop && !input.SwitchWeapon {
			continue
		}

		inv, invOk := world.Inventory.Get(entityID)
		pos, posOk := world.Position.Get(entityID)
		if !invOk || !posOk {
			continue
		}

		if input.PickupEntityID != state.NoPickup {
			is.handlePickup(entityID, pos, &inv, state.EntityID(input.PickupEntityID))
		}
		if input.DropSlotIndex != state.NoDrop {
			is.handleDrop(entityID, pos, &inv, input.DropSlotIndex)
		}
		if input.SwitchWeapon {
			is.handleWeaponSwitch(entityID, &inv)
		}

		world.UpdatePlayer(entityID, state.UpdatePlayer{
			UpdateMeta: state.ComponentInventory,
			Inventory:  inv,
		})
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
		inv.Weapons[slotIdx] = state.WeaponSlot{ItemDefID: gi.ItemDefID}
	} else {
		// Non-weapon item goes to item slots
		slotIdx := findFreeItemSlot(*inv)
		if slotIdx < 0 {
			return
		}
		inv.Items[slotIdx] = state.ItemSlot{
			ItemDefID: gi.ItemDefID,
			Quantity:  gi.Quantity,
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

	if slotIndex < 3 {
		// Weapon slot drop
		if slotIndex == 0 {
			return // cannot drop Fist
		}
		ws := inv.Weapons[slotIndex]
		if ws.IsEmpty() {
			return
		}
		itemDefID = ws.ItemDefID
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
		inv.Items[itemIdx] = state.ItemSlot{}
	}

	fwd := vector.Forward(float64(dir)).Scale(1.5)
	dropPos := vector.Vector2D(playerPos).Add(fwd)

	is.world.CreateGroundItemEntity(state.CreateGroundItem{
		Position:  state.Position(dropPos),
		ItemDefID: itemDefID,
		Quantity:  quantity,
	})
}

func (is *InventorySystem) handleWeaponSwitch(entityID state.EntityID, inv *state.Inventory) {
	tick := *is.currentTick

	if tick-inv.LastSwitchTick < weaponSwitchCooldownTicks {
		return
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
