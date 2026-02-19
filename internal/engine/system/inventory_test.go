package system

import (
	"math"
	"testing"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
)

func setupInventoryWorld() *state.World {
	w := state.NewWorld(5.0, 20, 20)
	// Reserve entity 0 so no real entity gets EntityID(0),
	// which would be confused with "empty" in IsEmpty() checks.
	w.Entity.Alloc()
	return w
}

func createFistDef(t *testing.T, world *state.World) state.EntityID {
	t.Helper()
	id, ok := world.CreateItemDefEntity(state.ItemConfig{
		Name:         "Fist",
		Type:         state.ItemTypeWeapon,
		MaxStack:     1,
		WeaponConfig: state.FistSpec,
	})
	if !ok {
		t.Fatal("failed to create fist def entity")
	}
	return id
}

func createItemDef(t *testing.T, world *state.World, name string, itemType state.ItemType, maxStack int) state.EntityID {
	t.Helper()
	id, ok := world.CreateItemDefEntity(state.ItemConfig{
		Name:     name,
		Type:     itemType,
		MaxStack: maxStack,
	})
	if !ok {
		t.Fatal("failed to create item def entity")
	}
	return id
}

func createWeaponItemDef(t *testing.T, world *state.World, name string, spec state.WeaponConfig) state.EntityID {
	t.Helper()
	id, ok := world.CreateItemDefEntity(state.ItemConfig{
		Name:         name,
		Type:         state.ItemTypeWeapon,
		MaxStack:     1,
		WeaponConfig: spec,
	})
	if !ok {
		t.Fatal("failed to create weapon item def entity")
	}
	return id
}

func createGroundItem(t *testing.T, world *state.World, pos state.Position, itemDefID state.EntityID, quantity int) state.EntityID {
	t.Helper()
	id, ok := world.CreateGroundItemEntity(state.CreateGroundItem{
		Position:  pos,
		ItemDefID: itemDefID,
		Quantity:  quantity,
	})
	if !ok {
		t.Fatal("failed to create ground item")
	}
	world.ApplyCommands()
	return id
}

func createPlayerForInventory(t *testing.T, world *state.World, pos state.Position, dir state.Direction, fistDefID state.EntityID) state.EntityID {
	t.Helper()
	id, ok := world.CreatePlayer(state.CreatePlayer{
		Position:      pos,
		Direction:     dir,
		MovementSpeed: 5.0,
		RotationSpeed: 2.0,
		Radius:        0.5,
		Health:        100,
		FistDefID:     fistDefID,
	})
	if !ok {
		t.Fatal("failed to create player")
	}
	world.ApplyCommands()
	return id
}

// --- Item pickup tests ---

func TestPickup_BasicItem(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	materialDef := createItemDef(t, world, "Wood", state.ItemTypeMaterial, 64)

	itemPos := state.Position{X: 10.5, Y: 10.5}
	itemID := createGroundItem(t, world, itemPos, materialDef, 5)

	world.SetInput(playerID, state.Input{PickupEntityID: int64(itemID), DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()

	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, ok := world.Inventory.Get(playerID)
	if !ok {
		t.Fatal("player should have Inventory component")
	}
	if inv.Items[0].ItemDefID != materialDef {
		t.Errorf("Expected item slot 0 to have item def %d, got %d", materialDef, inv.Items[0].ItemDefID)
	}
	if inv.Items[0].Quantity != 5 {
		t.Errorf("Expected quantity 5, got %d", inv.Items[0].Quantity)
	}

	if world.Entity.IsAlive(itemID) {
		t.Error("ground item entity should be destroyed after pickup")
	}
}

func TestPickup_WeaponRouting_KnifeToSlot1(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	knifeDef := createWeaponItemDef(t, world, "Knife", state.WeaponConfig{
		Type: state.WeaponTypeKnife, Range: 1, FireRate: 7, Damage: 15, Speed: 30,
	})
	knifeItemID := createGroundItem(t, world, state.Position{X: 10.3, Y: 10}, knifeDef, 1)

	world.SetInput(playerID, state.Input{PickupEntityID: int64(knifeItemID), DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()

	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(playerID)
	if inv.Weapons[1].ItemDefID != knifeDef {
		t.Errorf("Expected knife in weapon slot 1, got ItemDefID %d", inv.Weapons[1].ItemDefID)
	}
}

func TestPickup_WeaponRouting_GunToSlot2(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	gunDef := createWeaponItemDef(t, world, "Pistol", state.WeaponConfig{
		Type: state.WeaponTypeGun, Range: 20, FireRate: 2, Damage: 30, Speed: 30,
	})
	gunItemID := createGroundItem(t, world, state.Position{X: 10.3, Y: 10}, gunDef, 1)

	world.SetInput(playerID, state.Input{PickupEntityID: int64(gunItemID), DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()

	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(playerID)
	if inv.Weapons[2].ItemDefID != gunDef {
		t.Errorf("Expected gun in weapon slot 2, got ItemDefID %d", inv.Weapons[2].ItemDefID)
	}
}

func TestPickup_OutOfRange(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	materialDef := createItemDef(t, world, "FarItem", state.ItemTypeMaterial, 1)
	itemID := createGroundItem(t, world, state.Position{X: 20, Y: 20}, materialDef, 1)

	world.SetInput(playerID, state.Input{PickupEntityID: int64(itemID), DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()

	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(playerID)
	if !inv.Items[0].IsEmpty() {
		t.Error("item slot 0 should be empty when item is out of range")
	}

	_, groundOk := world.GroundItem.Get(itemID)
	if !groundOk {
		t.Error("item should still have GroundItem component")
	}
}

func TestPickup_FullItems(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	// Fill all item slots
	inv, _ := world.Inventory.Get(playerID)
	for i := 0; i < ports.ItemSlotCount; i++ {
		fillerDef := createItemDef(t, world, "filler", state.ItemTypeMaterial, 1)
		inv.Items[i] = state.ItemSlot{ItemDefID: fillerDef, Quantity: 1}
	}
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	extraDef := createItemDef(t, world, "ExtraItem", state.ItemTypeMaterial, 1)
	extraItemID := createGroundItem(t, world, state.Position{X: 10.5, Y: 10}, extraDef, 1)

	world.SetInput(playerID, state.Input{PickupEntityID: int64(extraItemID), DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()

	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	_, groundOk := world.GroundItem.Get(extraItemID)
	if !groundOk {
		t.Error("item should still be on ground when inventory is full")
	}
}

func TestPickup_ExplicitTargeting(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	farDef := createItemDef(t, world, "FarItem", state.ItemTypeMaterial, 1)
	nearDef := createItemDef(t, world, "NearItem", state.ItemTypeMaterial, 1)

	farItemID := createGroundItem(t, world, state.Position{X: 11.5, Y: 10}, farDef, 1)
	createGroundItem(t, world, state.Position{X: 10.3, Y: 10}, nearDef, 1)

	// Explicitly target the farther item (both within range)
	world.SetInput(playerID, state.Input{PickupEntityID: int64(farItemID), DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()

	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(playerID)
	if inv.Items[0].ItemDefID != farDef {
		t.Errorf("Expected targeted far item def %d in slot 0, got %d", farDef, inv.Items[0].ItemDefID)
	}
}

func TestPickup_InvalidEntityID(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	// Target a non-existent entity ID
	world.SetInput(playerID, state.Input{PickupEntityID: 9999, DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()

	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(playerID)
	for i := 0; i < ports.ItemSlotCount; i++ {
		if !inv.Items[i].IsEmpty() {
			t.Errorf("item slot %d should be empty when targeting invalid entity", i)
		}
	}
}

func TestPickup_NotGroundItem(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	// Create another player entity (not a ground item)
	otherPlayerID := createPlayerForInventory(t, world, state.Position{X: 10.5, Y: 10}, 0, fistDef)

	// Try to pick up the other player
	world.SetInput(playerID, state.Input{PickupEntityID: int64(otherPlayerID), DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()

	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(playerID)
	for i := 0; i < ports.ItemSlotCount; i++ {
		if !inv.Items[i].IsEmpty() {
			t.Errorf("item slot %d should be empty when targeting non-ground-item entity", i)
		}
	}

	// Other player should still be alive
	if !world.Entity.IsAlive(otherPlayerID) {
		t.Error("non-ground-item entity should not be destroyed")
	}
}

// --- Drop tests ---

func TestDrop_ItemSlot(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerDir := state.Direction(0)
	playerID := createPlayerForInventory(t, world, playerPos, playerDir, fistDef)

	materialDef := createItemDef(t, world, "Wood", state.ItemTypeMaterial, 64)

	inv, _ := world.Inventory.Get(playerID)
	inv.Items[0] = state.ItemSlot{ItemDefID: materialDef, Quantity: 3}
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	// Drop from item slot 0 -> DropSlotIndex = 3 (0 + 3)
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: 3})
	world.SyncInputBuffer()

	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if !inv.Items[0].IsEmpty() {
		t.Error("item slot 0 should be empty after drop")
	}

	var foundID state.EntityID
	for eid, gi := range world.GroundItem.All() {
		if gi.ItemDefID == materialDef {
			foundID = eid
			break
		}
	}
	if foundID == 0 {
		t.Fatal("dropped item should exist as a ground item")
	}

	gi, _ := world.GroundItem.Get(foundID)
	if gi.Quantity != 3 {
		t.Errorf("Expected ground item quantity 3, got %d", gi.Quantity)
	}

	itemPos, posOk := world.Position.Get(foundID)
	if !posOk {
		t.Fatal("item should have Position component after drop")
	}

	expectedX := 10.0
	expectedY := 10.0 - 1.5
	if math.Abs(itemPos.X-expectedX) > 1e-6 || math.Abs(itemPos.Y-expectedY) > 1e-6 {
		t.Errorf("Expected drop position (%.2f, %.2f), got (%.2f, %.2f)", expectedX, expectedY, itemPos.X, itemPos.Y)
	}
}

func TestDrop_WeaponSlot(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	knifeDef := createWeaponItemDef(t, world, "Knife", state.WeaponConfig{
		Type: state.WeaponTypeKnife, Range: 1, FireRate: 7, Damage: 15, Speed: 30,
	})

	inv, _ := world.Inventory.Get(playerID)
	inv.Weapons[1] = state.WeaponSlot{ItemDefID: knifeDef}
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	// Drop weapon slot 1
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: 1})
	world.SyncInputBuffer()

	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if !inv.Weapons[1].IsEmpty() {
		t.Error("weapon slot 1 should be empty after drop")
	}

	var foundID state.EntityID
	for eid, gi := range world.GroundItem.All() {
		if gi.ItemDefID == knifeDef {
			foundID = eid
			break
		}
	}
	if foundID == 0 {
		t.Fatal("dropped weapon should exist as a ground item")
	}
}

func TestDrop_CannotDropFist(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	// Try to drop slot 0 (Fist)
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: 0})
	world.SyncInputBuffer()

	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(playerID)
	if inv.Weapons[0].IsEmpty() {
		t.Error("Fist (weapon slot 0) should not be droppable")
	}
}

func TestDrop_EmptySlot(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	// Drop from empty item slot
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: 3})
	world.SyncInputBuffer()

	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(playerID)
	if !inv.Items[0].IsEmpty() {
		t.Error("slot should remain empty")
	}
}

func TestDrop_InvalidSlotIndex(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 1
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	createPlayerForInventory(t, world, playerPos, 0, fistDef)

	// Drop with out-of-bounds index - should not crash
	world.SetInput(1, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: 99})
	world.SyncInputBuffer()
	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	// Negative
	world.SetInput(1, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()
	is.Update(ports.DeltaTime)
	world.ApplyCommands()
}

// --- Weapon switch tests ---

func TestWeaponSwitch_CyclesOccupiedSlots(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 100
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	// Add knife to slot 1
	knifeDef := createWeaponItemDef(t, world, "Knife", state.WeaponConfig{
		Type: state.WeaponTypeKnife, Range: 1, FireRate: 7, Damage: 15, Speed: 30,
	})
	inv, _ := world.Inventory.Get(playerID)
	inv.Weapons[1] = state.WeaponSlot{ItemDefID: knifeDef}
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	// Currently at slot 0 (Fist), switch -> slot 1 (Knife)
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, SwitchWeapon: true})
	world.SyncInputBuffer()
	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.CurrentWeaponIndex != 1 {
		t.Errorf("Expected weapon index 1 after first switch, got %d", inv.CurrentWeaponIndex)
	}

	// Advance past cooldown, switch -> back to slot 0 (Fist), skip empty slot 2
	tick += weaponSwitchCooldownTicks
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, SwitchWeapon: true})
	world.SyncInputBuffer()
	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.CurrentWeaponIndex != 0 {
		t.Errorf("Expected weapon index 0 after second switch (skip empty slot 2), got %d", inv.CurrentWeaponIndex)
	}
}

func TestWeaponSwitch_CooldownEnforced(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 100
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	// Add knife
	knifeDef := createWeaponItemDef(t, world, "Knife", state.WeaponConfig{
		Type: state.WeaponTypeKnife, Range: 1, FireRate: 7, Damage: 15, Speed: 30,
	})
	inv, _ := world.Inventory.Get(playerID)
	inv.Weapons[1] = state.WeaponSlot{ItemDefID: knifeDef}
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	// First switch: 0 -> 1
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, SwitchWeapon: true})
	world.SyncInputBuffer()
	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.CurrentWeaponIndex != 1 {
		t.Fatalf("Expected weapon index 1, got %d", inv.CurrentWeaponIndex)
	}

	// Try again immediately - should be blocked
	tick += 1
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, SwitchWeapon: true})
	world.SyncInputBuffer()
	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.CurrentWeaponIndex != 1 {
		t.Errorf("Expected weapon index 1 (cooldown blocked), got %d", inv.CurrentWeaponIndex)
	}
}

func TestWeaponSwitch_SkipsEmptySlots(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 100
	fistDef := createFistDef(t, world)
	is := NewInventorySystem(world, &tick)

	playerPos := state.Position{X: 10, Y: 10}
	playerID := createPlayerForInventory(t, world, playerPos, 0, fistDef)

	// Add gun to slot 2 (slot 1 remains empty)
	gunDef := createWeaponItemDef(t, world, "Pistol", state.WeaponConfig{
		Type: state.WeaponTypeGun, Range: 20, FireRate: 2, Damage: 30, Speed: 30,
	})
	inv, _ := world.Inventory.Get(playerID)
	inv.Weapons[2] = state.WeaponSlot{ItemDefID: gunDef}
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	// Switch from 0 -> should skip empty 1 -> go to 2
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, SwitchWeapon: true})
	world.SyncInputBuffer()
	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.CurrentWeaponIndex != 2 {
		t.Errorf("Expected weapon index 2 (skip empty slot 1), got %d", inv.CurrentWeaponIndex)
	}

	// Switch from 2 -> back to 0
	tick += weaponSwitchCooldownTicks
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, SwitchWeapon: true})
	world.SyncInputBuffer()
	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.CurrentWeaponIndex != 0 {
		t.Errorf("Expected weapon index 0, got %d", inv.CurrentWeaponIndex)
	}
}
