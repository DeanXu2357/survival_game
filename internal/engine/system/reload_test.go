package system

import (
	"testing"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
)

// createPistolMagDef creates an ItemConfig entity for a pistol magazine.
func createPistolMagDef(t *testing.T, world *state.World, magCapacity int) state.EntityID {
	t.Helper()
	id, ok := world.CreateItemDefEntity(state.ItemConfig{
		Name:         "PistolMag",
		Type:         state.ItemTypeMagazine,
		MaxStack:     1,
		MagCapacity:  magCapacity,
		AmmoCategory: state.AmmoCategoryPistol,
	})
	if !ok {
		t.Fatal("failed to create pistol mag def")
	}
	return id
}

// createGunDefWithAmmo creates a gun weapon def that uses AmmoCategoryPistol.
func createGunDefWithAmmo(t *testing.T, world *state.World) state.EntityID {
	t.Helper()
	id, ok := world.CreateItemDefEntity(state.ItemConfig{
		Name:     "Pistol",
		Type:     state.ItemTypeWeapon,
		MaxStack: 1,
		WeaponConfig: state.WeaponConfig{
			Type:         state.WeaponTypeGun,
			Range:        20,
			FireRate:     2,
			Damage:       30,
			Speed:        30,
			AmmoCategory: state.AmmoCategoryPistol,
		},
	})
	if !ok {
		t.Fatal("failed to create gun def with ammo")
	}
	return id
}

// setupReloadPlayer creates a world, player with a pistol (loaded with a magazine) and spare mags.
func setupReloadPlayer(t *testing.T) (world *state.World, playerID state.EntityID, gunDef state.EntityID, magDef state.EntityID) {
	t.Helper()
	world = setupInventoryWorld()

	fistDef := createFistDef(t, world)
	gunDef = createGunDefWithAmmo(t, world)
	magDef = createPistolMagDef(t, world, 12)

	playerID = createPlayerForInventory(t, world, state.Position{X: 50, Y: 50}, 0, fistDef)

	// Equip gun in slot 2 with a loaded magazine (8/12 ammo)
	inv, _ := world.Inventory.Get(playerID)
	inv.Weapons[2] = state.WeaponSlot{
		ItemID:        gunDef,
		LoadedMagID:   magDef,
		LoadedMagAmmo: 8,
		MagCapacity:   12,
	}
	inv.CurrentWeaponIndex = 2

	// Put a spare magazine in item slot 0 (12/12 ammo)
	inv.Items[0] = state.ItemSlot{ItemDefID: magDef, Quantity: 1, Ammo: 12}

	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	return world, playerID, gunDef, magDef
}

func TestReload_NormalReloadCompletesAfter90Ticks(t *testing.T) {
	world, playerID, _, magDef := setupReloadPlayer(t)
	var tick uint64 = 100

	rs := NewInventorySystem(world, &tick)

	// Start normal reload
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Reload: true})
	world.SyncInputBuffer()
	rs.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(playerID)
	if inv.Weapons[2].ReloadStartTick != 100 {
		t.Fatalf("Expected ReloadStartTick=100, got %d", inv.Weapons[2].ReloadStartTick)
	}
	if inv.Weapons[2].ReloadType != state.ReloadTypeNormal {
		t.Fatalf("Expected ReloadTypeNormal, got %d", inv.Weapons[2].ReloadType)
	}

	// Advance to tick 189 (89 ticks elapsed) - reload should NOT be complete
	tick = 189
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()
	rs.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.Weapons[2].ReloadStartTick == 0 {
		t.Fatal("Reload should not be complete at tick 189 (only 89 ticks elapsed)")
	}

	// Advance to tick 190 (90 ticks elapsed) - reload should complete
	tick = 190
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()
	rs.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.Weapons[2].ReloadStartTick != 0 {
		t.Errorf("Expected ReloadStartTick=0 after completion, got %d", inv.Weapons[2].ReloadStartTick)
	}
	if inv.Weapons[2].ReloadType != state.ReloadTypeNone {
		t.Errorf("Expected ReloadTypeNone after completion, got %d", inv.Weapons[2].ReloadType)
	}
	// New magazine should be loaded (12/12 from spare)
	if inv.Weapons[2].LoadedMagAmmo != 12 {
		t.Errorf("Expected LoadedMagAmmo=12, got %d", inv.Weapons[2].LoadedMagAmmo)
	}
	if inv.Weapons[2].LoadedMagID != magDef {
		t.Errorf("Expected LoadedMagID=%d, got %d", magDef, inv.Weapons[2].LoadedMagID)
	}

	// Old magazine (8 rounds) should be back in inventory
	found := false
	for _, slot := range inv.Items {
		if slot.ItemDefID == magDef && slot.Ammo == 8 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Old magazine with 8 rounds should be returned to inventory on normal reload")
	}

	// Spare magazine slot should be cleared
	if inv.Items[0].ItemDefID == magDef && inv.Items[0].Ammo == 12 {
		t.Error("Spare magazine should have been consumed from inventory")
	}
}

func TestReload_FastReloadCompletesAfter60Ticks(t *testing.T) {
	world, playerID, _, magDef := setupReloadPlayer(t)
	var tick uint64 = 100

	rs := NewInventorySystem(world, &tick)

	// Start fast reload
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, FastReload: true})
	world.SyncInputBuffer()
	rs.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(playerID)
	if inv.Weapons[2].ReloadType != state.ReloadTypeFast {
		t.Fatalf("Expected ReloadTypeFast, got %d", inv.Weapons[2].ReloadType)
	}

	// Advance to tick 159 (59 ticks) - should NOT complete
	tick = 159
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()
	rs.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.Weapons[2].ReloadStartTick == 0 {
		t.Fatal("Reload should not be complete at tick 159")
	}

	// Advance to tick 160 (60 ticks) - should complete
	tick = 160
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()
	rs.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.Weapons[2].ReloadStartTick != 0 {
		t.Errorf("Expected ReloadStartTick=0 after fast reload, got %d", inv.Weapons[2].ReloadStartTick)
	}
	if inv.Weapons[2].LoadedMagAmmo != 12 {
		t.Errorf("Expected LoadedMagAmmo=12, got %d", inv.Weapons[2].LoadedMagAmmo)
	}

	// Fast reload drops old mag as GroundItem (8 rounds)
	foundGround := false
	for _, gi := range world.GroundItem.All() {
		if gi.ItemDefID == magDef && gi.Ammo == 8 {
			foundGround = true
			break
		}
	}
	if !foundGround {
		t.Error("Old magazine should be dropped as ground item on fast reload")
	}

	// Old mag should NOT be in inventory
	for _, slot := range inv.Items {
		if slot.ItemDefID == magDef && slot.Ammo == 8 {
			t.Error("Old magazine should NOT be in inventory after fast reload")
			break
		}
	}
}

func TestReload_CannotFireDuringReload(t *testing.T) {
	world, playerID, _, _ := setupReloadPlayer(t)
	var tick uint64 = 100

	is := NewInventorySystem(world, &tick)
	wf := NewWeaponFireSystem(world, &tick)

	// Start reload
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Reload: true})
	world.SyncInputBuffer()
	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	// Try to fire during reload
	tick = 110
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()
	is.Update(ports.DeltaTime)
	wf.Update(ports.DeltaTime)
	world.ApplyCommands()

	projectileCount := 0
	for range world.Projectile.All() {
		projectileCount++
	}
	if projectileCount != 0 {
		t.Errorf("Expected 0 projectiles during reload, got %d", projectileCount)
	}
}

func TestReload_WeaponSwitchCancelsReload(t *testing.T) {
	world, playerID, _, _ := setupReloadPlayer(t)
	var tick uint64 = 100

	is := NewInventorySystem(world, &tick)

	// Add knife so we can switch
	knifeDef := createWeaponItemDef(t, world, "Knife", state.WeaponConfig{
		Type: state.WeaponTypeKnife, Range: 1, FireRate: 7, Damage: 15, Speed: 30,
	})
	inv, _ := world.Inventory.Get(playerID)
	inv.Weapons[1] = state.WeaponSlot{ItemID: knifeDef}
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	// Start reload
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Reload: true})
	world.SyncInputBuffer()
	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.Weapons[2].ReloadStartTick == 0 {
		t.Fatal("Reload should have started")
	}

	// Switch weapon - should cancel reload
	tick = 110
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, SwitchWeapon: true})
	world.SyncInputBuffer()
	is.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.Weapons[2].ReloadStartTick != 0 {
		t.Errorf("Reload on gun should be cancelled after weapon switch, got ReloadStartTick=%d", inv.Weapons[2].ReloadStartTick)
	}
	if inv.Weapons[2].ReloadType != state.ReloadTypeNone {
		t.Errorf("ReloadType should be None after cancel, got %d", inv.Weapons[2].ReloadType)
	}
}

func TestReload_CannotReloadWithoutSpareMagazine(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 100

	fistDef := createFistDef(t, world)
	gunDef := createGunDefWithAmmo(t, world)
	magDef := createPistolMagDef(t, world, 12)

	playerID := createPlayerForInventory(t, world, state.Position{X: 50, Y: 50}, 0, fistDef)

	// Equip gun with loaded mag but NO spare magazines
	inv, _ := world.Inventory.Get(playerID)
	inv.Weapons[2] = state.WeaponSlot{
		ItemID:        gunDef,
		LoadedMagID:   magDef,
		LoadedMagAmmo: 5,
		MagCapacity:   12,
	}
	inv.CurrentWeaponIndex = 2
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	rs := NewInventorySystem(world, &tick)

	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Reload: true})
	world.SyncInputBuffer()
	rs.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.Weapons[2].ReloadStartTick != 0 {
		t.Error("Should not start reload without spare magazines")
	}
}

func TestReload_CannotReloadWhenMagazineFull(t *testing.T) {
	world, playerID, _, magDef := setupReloadPlayer(t)
	var tick uint64 = 100

	// Set loaded mag to full (12/12)
	inv, _ := world.Inventory.Get(playerID)
	inv.Weapons[2].LoadedMagAmmo = 12
	inv.Weapons[2].MagCapacity = 12
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()
	_ = magDef

	rs := NewInventorySystem(world, &tick)

	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Reload: true})
	world.SyncInputBuffer()
	rs.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.Weapons[2].ReloadStartTick != 0 {
		t.Error("Should not start reload when magazine is already full")
	}
}

func TestReload_FullestMagazineSelectedFirst(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 100

	fistDef := createFistDef(t, world)
	gunDef := createGunDefWithAmmo(t, world)
	magDef := createPistolMagDef(t, world, 12)

	playerID := createPlayerForInventory(t, world, state.Position{X: 50, Y: 50}, 0, fistDef)

	inv, _ := world.Inventory.Get(playerID)
	inv.Weapons[2] = state.WeaponSlot{
		ItemID:        gunDef,
		LoadedMagID:   magDef,
		LoadedMagAmmo: 2,
		MagCapacity:   12,
	}
	inv.CurrentWeaponIndex = 2

	// Put magazines with different ammo counts
	inv.Items[0] = state.ItemSlot{ItemDefID: magDef, Quantity: 1, Ammo: 5}  // 5 rounds
	inv.Items[1] = state.ItemSlot{ItemDefID: magDef, Quantity: 1, Ammo: 10} // 10 rounds (fullest)
	inv.Items[2] = state.ItemSlot{ItemDefID: magDef, Quantity: 1, Ammo: 3}  // 3 rounds

	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	rs := NewInventorySystem(world, &tick)

	// Start normal reload
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Reload: true})
	world.SyncInputBuffer()
	rs.Update(ports.DeltaTime)
	world.ApplyCommands()

	// Complete reload
	tick += ports.NormalReloadTicks
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop})
	world.SyncInputBuffer()
	rs.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.Weapons[2].LoadedMagAmmo != 10 {
		t.Errorf("Expected fullest magazine (10 rounds) to be loaded, got %d", inv.Weapons[2].LoadedMagAmmo)
	}

	// The slot that had 10 rounds should now be empty
	if !inv.Items[1].IsEmpty() {
		t.Errorf("Item slot 1 (fullest mag) should be empty after reload, got ammo=%d", inv.Items[1].Ammo)
	}
}

func TestReload_FireDeductsAmmoAndBlocksAtZero(t *testing.T) {
	world, playerID, _, _ := setupReloadPlayer(t)
	var tick uint64 = 100

	// Set loaded mag to 2 rounds
	inv, _ := world.Inventory.Get(playerID)
	inv.Weapons[2].LoadedMagAmmo = 2
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	wf := NewWeaponFireSystem(world, &tick)

	// Fire first shot
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()
	wf.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.Weapons[2].LoadedMagAmmo != 1 {
		t.Errorf("Expected 1 ammo after first shot, got %d", inv.Weapons[2].LoadedMagAmmo)
	}

	// Fire second shot (need to advance past fire rate: gun FireRate=2 -> interval=30 ticks)
	tick += 30
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()
	wf.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ = world.Inventory.Get(playerID)
	if inv.Weapons[2].LoadedMagAmmo != 0 {
		t.Errorf("Expected 0 ammo after second shot, got %d", inv.Weapons[2].LoadedMagAmmo)
	}

	projectileCount := 0
	for range world.Projectile.All() {
		projectileCount++
	}
	if projectileCount != 2 {
		t.Fatalf("Expected 2 projectiles after 2 shots, got %d", projectileCount)
	}

	// Third shot should be blocked (0 ammo)
	tick += 30
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()
	wf.Update(ports.DeltaTime)
	world.ApplyCommands()

	projectileCount = 0
	for range world.Projectile.All() {
		projectileCount++
	}
	if projectileCount != 2 {
		t.Errorf("Expected still 2 projectiles (no ammo), got %d", projectileCount)
	}
}

func TestReload_MeleeWeaponCannotReload(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 100

	fistDef := createFistDef(t, world)
	playerID := createPlayerForInventory(t, world, state.Position{X: 50, Y: 50}, 0, fistDef)

	rs := NewInventorySystem(world, &tick)

	// Try reload with fist equipped (slot 0)
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Reload: true})
	world.SyncInputBuffer()
	rs.Update(ports.DeltaTime)
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(playerID)
	if inv.Weapons[0].ReloadStartTick != 0 {
		t.Error("Melee weapon should not start reload")
	}
}

func TestReload_MeleeFiresWithoutAmmo(t *testing.T) {
	world := setupInventoryWorld()
	var tick uint64 = 100

	fistDef := createFistDef(t, world)
	playerID := createPlayerForInventory(t, world, state.Position{X: 50, Y: 50}, 0, fistDef)

	wf := NewWeaponFireSystem(world, &tick)

	// Fist should fire without ammo (AmmoCategoryNone)
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()
	wf.Update(ports.DeltaTime)
	world.ApplyCommands()

	projectileCount := 0
	for range world.Projectile.All() {
		projectileCount++
	}
	if projectileCount != 1 {
		t.Errorf("Expected 1 projectile for fist (no ammo needed), got %d", projectileCount)
	}
}
