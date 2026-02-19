package state

import "survival/internal/engine/ports"

// ItemSlot represents a non-weapon item in the player's inventory.
type ItemSlot struct {
	ItemDefID EntityID
	Quantity  int
}

func (s ItemSlot) IsEmpty() bool { return s.ItemDefID == 0 }

// Inventory is the unified component replacing Backpack + WeaponState.
// Weapons[0]=Fist (always), Weapons[1]=Knife slot, Weapons[2]=Gun slot.
type Inventory struct {
	Weapons            [3]WeaponSlot
	Items              [ports.ItemSlotCount]ItemSlot
	CurrentWeaponIndex int
	LastSwitchTick     uint64
	MaxItemCapacity    int
}

// ItemCount returns the number of occupied item slots.
func (inv Inventory) ItemCount() int {
	n := 0
	for _, s := range inv.Items {
		if !s.IsEmpty() {
			n++
		}
	}
	return n
}

// IsItemsFull returns true if all item slots up to MaxItemCapacity are occupied.
func (inv Inventory) IsItemsFull() bool {
	return inv.ItemCount() >= inv.MaxItemCapacity
}

// OccupiedWeaponSlots returns indices of non-empty weapon slots.
func (inv Inventory) OccupiedWeaponSlots() []int {
	var slots []int
	for i, ws := range inv.Weapons {
		if !ws.IsEmpty() {
			slots = append(slots, i)
		}
	}
	return slots
}

type ItemType uint8

const (
	ItemTypeWeapon ItemType = iota
	ItemTypeConsumable
	ItemTypeMaterial
	ItemTypeEquipment
)

type ItemConfig struct {
	Name         string
	Type         ItemType
	MaxStack     int
	WeaponConfig WeaponConfig // zero-valued for non-weapons
}

type GroundItem struct {
	ItemDefID EntityID
	Quantity  int
}
