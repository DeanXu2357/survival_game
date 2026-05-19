package engine_test

import (
	"math"
	"testing"

	"survival/internal/engine"
	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/vector"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestGame(t *testing.T, walls []engine.WallConfig, spawnX, spawnY float64) *engine.Game {
	t.Helper()
	mapConfig := &engine.MapConfig{
		Dimensions:  vector.Vector2D{X: 200, Y: 200},
		GridSize:    10,
		Walls:       walls,
		SpawnPoints: []engine.SpawnPoint{{Position: vector.Vector2D{X: spawnX, Y: spawnY}}},
	}
	game, err := engine.NewGame(mapConfig)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	game.StartGameLoop()
	return game
}

func noopInput() ports.PlayerInput {
	return ports.PlayerInput{PickupEntityID: ports.NoPickup, DropSlotIndex: ports.NoDrop}
}

func tickN(t *testing.T, g *engine.Game, pid state.EntityID, n int) {
	t.Helper()
	inp := noopInput()
	for i := 0; i < n; i++ {
		g.SetPlayerInput(pid, inp)
		g.Update(ports.DeltaTime)
	}
}

func tickWithInput(t *testing.T, g *engine.Game, pid state.EntityID, inp ports.PlayerInput, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		g.SetPlayerInput(pid, inp)
		g.Update(ports.DeltaTime)
	}
}

// gunSetup holds IDs returned by registerGunAndMag / joinAndEquipGun.
type gunSetup struct {
	playerID state.EntityID
	gunDefID state.EntityID
	magDefID state.EntityID
}

var testGunWeaponCfg = state.WeaponConfig{
	Type:         state.WeaponTypeGun,
	Range:        50,
	FireRate:     2, // fireInterval = 30 ticks
	Damage:       25,
	Speed:        30,
	AmmoCategory: state.AmmoCategoryPistol,
}

func registerGunAndMag(t *testing.T, g *engine.Game) (gunDefID, magDefID state.EntityID) {
	t.Helper()
	var err error
	gunDefID, err = g.RegisterItemDef(state.ItemConfig{
		Name:         "Pistol",
		Type:         state.ItemTypeWeapon,
		MaxStack:     1,
		WeaponConfig: testGunWeaponCfg,
	})
	if err != nil {
		t.Fatalf("RegisterItemDef gun: %v", err)
	}
	magDefID, err = g.RegisterItemDef(state.ItemConfig{
		Name:         "PistolMag",
		Type:         state.ItemTypeMagazine,
		MaxStack:     1,
		MagCapacity:  12,
		AmmoCategory: state.AmmoCategoryPistol,
	})
	if err != nil {
		t.Fatalf("RegisterItemDef mag: %v", err)
	}
	return
}

// joinAndEquipGun joins a player and equips a gun with magAmmo rounds loaded.
func joinAndEquipGun(t *testing.T, g *engine.Game, magAmmo int) gunSetup {
	t.Helper()
	pid, err := g.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}
	gunDefID, magDefID := registerGunAndMag(t, g)
	inv, ok := g.PlayerInventory(pid)
	if !ok {
		t.Fatal("PlayerInventory not found")
	}
	inv.Weapons[2] = state.WeaponSlot{
		ItemID:        gunDefID,
		LoadedMagID:   magDefID,
		LoadedMagAmmo: magAmmo,
		MagCapacity:   12,
	}
	inv.CurrentWeaponIndex = 2
	if err := g.SetPlayerInventory(pid, inv); err != nil {
		t.Fatalf("SetPlayerInventory: %v", err)
	}
	return gunSetup{playerID: pid, gunDefID: gunDefID, magDefID: magDefID}
}

func registerKnifeDef(t *testing.T, g *engine.Game) state.EntityID {
	t.Helper()
	id, err := g.RegisterItemDef(state.ItemConfig{
		Name:     "Knife",
		Type:     state.ItemTypeWeapon,
		MaxStack: 1,
		WeaponConfig: state.WeaponConfig{
			Type:     state.WeaponTypeKnife,
			Range:    3,
			FireRate: 2,
			Damage:   30,
			Speed:    40,
		},
	})
	if err != nil {
		t.Fatalf("RegisterItemDef knife: %v", err)
	}
	return id
}

// ---------------------------------------------------------------------------
// Movement
// ---------------------------------------------------------------------------

func TestIntegration_MoveHorizontal(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	initial, _ := game.PlayerSnapshotWithLocation(pid)

	inp := noopInput()
	inp.MoveHorizontal = 1.0
	tickWithInput(t, game, pid, inp, 60) // 1 second at speed 5

	final, _ := game.PlayerSnapshotWithLocation(pid)
	dx := final.Player.Position.X - initial.Player.Position.X
	if math.Abs(dx-5.0) > 0.15 {
		t.Errorf("Expected X displacement ~5.0, got %.4f", dx)
	}
}

func TestIntegration_MoveRelative(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	// Rotate to face east (dir = pi/2) in 1 tick.
	// rotDelta = LookHorizontal * rotSpeed * dt = LH * 2 * (1/60) = LH/30
	// Need pi/2 → LH = pi/2 * 30 = 15*pi
	rotInp := noopInput()
	rotInp.LookHorizontal = math.Pi / 2 * 30
	tickWithInput(t, game, pid, rotInp, 1)

	snapBefore, _ := game.PlayerSnapshotWithLocation(pid)

	// Move forward in relative mode (MoveVertical=1 = forward along facing dir)
	moveInp := noopInput()
	moveInp.MoveVertical = 1.0
	moveInp.MovementType = ports.MovementTypeRelative
	tickWithInput(t, game, pid, moveInp, 60)

	snapAfter, _ := game.PlayerSnapshotWithLocation(pid)

	dx := snapAfter.Player.Position.X - snapBefore.Player.Position.X
	dy := math.Abs(snapAfter.Player.Position.Y - snapBefore.Player.Position.Y)
	if dx < 4.5 {
		t.Errorf("Expected X increase ~5.0 (facing east), got dx=%.4f", dx)
	}
	if dy > 0.5 {
		t.Errorf("Expected Y roughly unchanged, got dy=%.4f", dy)
	}
}

func TestIntegration_WallBlocksMovement(t *testing.T) {
	walls := []engine.WallConfig{{
		Center:   vector.Vector2D{X: 60, Y: 50},
		HalfSize: vector.Vector2D{X: 1, Y: 50},
	}}
	game := newTestGame(t, walls, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	inp := noopInput()
	inp.MoveHorizontal = 1.0
	tickWithInput(t, game, pid, inp, 180) // 3 seconds

	snap, _ := game.PlayerSnapshotWithLocation(pid)
	// Wall left edge at X=59, player radius=0.5 → max X ~58.5
	if snap.Player.Position.X > 59.1 {
		t.Errorf("Player walked through wall, X=%.4f", snap.Player.Position.X)
	}
	if snap.Player.Position.X < 57.0 {
		t.Errorf("Player didn't reach wall, X=%.4f", snap.Player.Position.X)
	}
}

// ---------------------------------------------------------------------------
// Weapon Fire
// ---------------------------------------------------------------------------

func TestIntegration_FistFireCreatesProjectile(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	inp := noopInput()
	inp.Fire = true
	tickWithInput(t, game, pid, inp, 1)

	if n := game.ProjectileCount(); n != 1 {
		t.Errorf("Expected 1 projectile, got %d", n)
	}
}

func TestIntegration_GunFireDeductsAmmo(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	setup := joinAndEquipGun(t, game, 12)

	inp := noopInput()
	inp.Fire = true
	tickWithInput(t, game, setup.playerID, inp, 1)

	inv, ok := game.PlayerInventory(setup.playerID)
	if !ok {
		t.Fatal("inventory not found")
	}
	if inv.Weapons[2].LoadedMagAmmo != 11 {
		t.Errorf("Expected LoadedMagAmmo=11, got %d", inv.Weapons[2].LoadedMagAmmo)
	}
}

func TestIntegration_GunFireRateLimit(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	setup := joinAndEquipGun(t, game, 12)

	inp := noopInput()
	inp.Fire = true
	// First fire
	tickWithInput(t, game, setup.playerID, inp, 1)
	if n := game.ProjectileCount(); n != 1 {
		t.Fatalf("Expected 1 projectile after first fire, got %d", n)
	}
	// Second fire next tick (within 30-tick interval) → should be blocked
	tickWithInput(t, game, setup.playerID, inp, 1)
	if n := game.ProjectileCount(); n != 1 {
		t.Errorf("Expected still 1 projectile (rate limited), got %d", n)
	}
}

func TestIntegration_GunCannotFireWithEmptyMag(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	setup := joinAndEquipGun(t, game, 0) // empty mag

	inp := noopInput()
	inp.Fire = true
	tickWithInput(t, game, setup.playerID, inp, 1)

	if n := game.ProjectileCount(); n != 0 {
		t.Errorf("Expected 0 projectiles with empty mag, got %d", n)
	}
}

// ---------------------------------------------------------------------------
// Reload
// ---------------------------------------------------------------------------

// addSpareMag adds a magazine to the first free item slot.
func addSpareMag(t *testing.T, g *engine.Game, pid, magDefID state.EntityID, ammo int) {
	t.Helper()
	inv, ok := g.PlayerInventory(pid)
	if !ok {
		t.Fatal("inventory not found")
	}
	for i := range inv.Items {
		if inv.Items[i].IsEmpty() {
			inv.Items[i] = state.ItemSlot{ItemDefID: magDefID, Quantity: 1, Ammo: ammo}
			if err := g.SetPlayerInventory(pid, inv); err != nil {
				t.Fatalf("SetPlayerInventory: %v", err)
			}
			return
		}
	}
	t.Fatal("no free item slot")
}

func TestIntegration_NormalReloadCompleteAt90Ticks(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	setup := joinAndEquipGun(t, game, 5) // 5/12 loaded
	addSpareMag(t, game, setup.playerID, setup.magDefID, 10)

	// Start normal reload (tick 1)
	inp := noopInput()
	inp.Reload = true
	tickWithInput(t, game, setup.playerID, inp, 1)

	inv, _ := game.PlayerInventory(setup.playerID)
	if inv.Weapons[2].ReloadStartTick == 0 {
		t.Fatal("Reload should have started")
	}

	// At T+89 (89 ticks later → tick 90): still reloading
	tickN(t, game, setup.playerID, 89)
	inv, _ = game.PlayerInventory(setup.playerID)
	if inv.Weapons[2].ReloadStartTick == 0 {
		t.Error("Reload should still be in progress at T+89")
	}

	// At T+90 (1 more → tick 91): complete
	tickN(t, game, setup.playerID, 1)
	inv, _ = game.PlayerInventory(setup.playerID)
	if inv.Weapons[2].ReloadStartTick != 0 {
		t.Error("Reload should be complete at T+90")
	}
	if inv.Weapons[2].LoadedMagAmmo != 10 {
		t.Errorf("Expected LoadedMagAmmo=10, got %d", inv.Weapons[2].LoadedMagAmmo)
	}
}

func TestIntegration_FastReloadCompleteAt60Ticks(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	setup := joinAndEquipGun(t, game, 5)
	addSpareMag(t, game, setup.playerID, setup.magDefID, 10)

	inp := noopInput()
	inp.FastReload = true
	tickWithInput(t, game, setup.playerID, inp, 1)

	inv, _ := game.PlayerInventory(setup.playerID)
	if inv.Weapons[2].ReloadStartTick == 0 {
		t.Fatal("Fast reload should have started")
	}

	// At T+59: still reloading
	tickN(t, game, setup.playerID, 59)
	inv, _ = game.PlayerInventory(setup.playerID)
	if inv.Weapons[2].ReloadStartTick == 0 {
		t.Error("Reload should still be in progress at T+59")
	}

	// At T+60: complete
	tickN(t, game, setup.playerID, 1)
	inv, _ = game.PlayerInventory(setup.playerID)
	if inv.Weapons[2].ReloadStartTick != 0 {
		t.Error("Fast reload should be complete at T+60")
	}
	if inv.Weapons[2].LoadedMagAmmo != 10 {
		t.Errorf("Expected LoadedMagAmmo=10, got %d", inv.Weapons[2].LoadedMagAmmo)
	}

	// Old magazine (5 ammo) should be dropped as ground item
	items := game.GroundItemSnapshots()
	if len(items) != 1 {
		t.Fatalf("Expected 1 ground item (dropped mag), got %d", len(items))
	}
	if items[0].GroundItem.Ammo != 5 {
		t.Errorf("Dropped mag should have 5 ammo, got %d", items[0].GroundItem.Ammo)
	}
}

func TestIntegration_CannotFireDuringReload(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	setup := joinAndEquipGun(t, game, 5)
	addSpareMag(t, game, setup.playerID, setup.magDefID, 10)

	// Start reload
	inp := noopInput()
	inp.Reload = true
	tickWithInput(t, game, setup.playerID, inp, 1)

	// Try to fire during reload
	fireInp := noopInput()
	fireInp.Fire = true
	tickWithInput(t, game, setup.playerID, fireInp, 1)

	if n := game.ProjectileCount(); n != 0 {
		t.Errorf("Should not fire during reload, got %d projectiles", n)
	}
}

func TestIntegration_WeaponSwitchCancelsReload(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	setup := joinAndEquipGun(t, game, 5)

	// Add knife so weapon switch has a target
	knifeDefID := registerKnifeDef(t, game)
	inv, _ := game.PlayerInventory(setup.playerID)
	inv.Weapons[1] = state.WeaponSlot{ItemID: knifeDefID}
	inv.Items[0] = state.ItemSlot{ItemDefID: setup.magDefID, Quantity: 1, Ammo: 10}
	if err := game.SetPlayerInventory(setup.playerID, inv); err != nil {
		t.Fatalf("SetPlayerInventory: %v", err)
	}

	// Wait past weapon switch cooldown (tick 0 → need >= 15 ticks)
	tickN(t, game, setup.playerID, 16)

	// Start reload
	reloadInp := noopInput()
	reloadInp.Reload = true
	tickWithInput(t, game, setup.playerID, reloadInp, 1)

	inv, _ = game.PlayerInventory(setup.playerID)
	if inv.Weapons[2].ReloadStartTick == 0 {
		t.Fatal("Reload should have started")
	}

	// Switch weapon → cancels reload
	switchInp := noopInput()
	switchInp.SwitchWeapon = true
	tickWithInput(t, game, setup.playerID, switchInp, 1)

	inv, _ = game.PlayerInventory(setup.playerID)
	if inv.Weapons[2].ReloadStartTick != 0 {
		t.Errorf("Reload should be cancelled, ReloadStartTick=%d", inv.Weapons[2].ReloadStartTick)
	}
}

func TestIntegration_CannotReloadWhenFull(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	setup := joinAndEquipGun(t, game, 12) // 12/12 = full
	addSpareMag(t, game, setup.playerID, setup.magDefID, 10)

	inp := noopInput()
	inp.Reload = true
	tickWithInput(t, game, setup.playerID, inp, 1)

	inv, _ := game.PlayerInventory(setup.playerID)
	if inv.Weapons[2].ReloadStartTick != 0 {
		t.Errorf("Should not reload when full, ReloadStartTick=%d", inv.Weapons[2].ReloadStartTick)
	}
}

func TestIntegration_FullestMagSelectedFirst(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	setup := joinAndEquipGun(t, game, 0) // empty gun

	// Add 3 spare mags with ammo 5, 10, 3
	inv, _ := game.PlayerInventory(setup.playerID)
	inv.Items[0] = state.ItemSlot{ItemDefID: setup.magDefID, Quantity: 1, Ammo: 5}
	inv.Items[1] = state.ItemSlot{ItemDefID: setup.magDefID, Quantity: 1, Ammo: 10}
	inv.Items[2] = state.ItemSlot{ItemDefID: setup.magDefID, Quantity: 1, Ammo: 3}
	if err := game.SetPlayerInventory(setup.playerID, inv); err != nil {
		t.Fatalf("SetPlayerInventory: %v", err)
	}

	// Start normal reload and wait for completion (90+1 ticks)
	reloadInp := noopInput()
	reloadInp.Reload = true
	tickWithInput(t, game, setup.playerID, reloadInp, 1)
	tickN(t, game, setup.playerID, 90) // T+90 → complete

	inv, _ = game.PlayerInventory(setup.playerID)
	if inv.Weapons[2].LoadedMagAmmo != 10 {
		t.Errorf("Expected fullest mag (ammo=10) loaded, got %d", inv.Weapons[2].LoadedMagAmmo)
	}
}

// ---------------------------------------------------------------------------
// Inventory Pickup / Drop
// ---------------------------------------------------------------------------

func TestIntegration_PickupWeapon(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	gunDefID, _ := registerGunAndMag(t, game)
	groundID, err := game.SpawnGroundItem(state.Position{X: 50.5, Y: 50}, gunDefID, 1, 0)
	if err != nil {
		t.Fatalf("SpawnGroundItem: %v", err)
	}

	inp := noopInput()
	inp.PickupEntityID = int64(groundID)
	tickWithInput(t, game, pid, inp, 1)

	inv, _ := game.PlayerInventory(pid)
	if inv.Weapons[2].IsEmpty() {
		t.Error("Expected gun in weapon slot 2")
	}
	if inv.Weapons[2].ItemID != gunDefID {
		t.Errorf("Expected ItemID=%d, got %d", gunDefID, inv.Weapons[2].ItemID)
	}
	if len(game.GroundItemSnapshots()) != 0 {
		t.Errorf("Ground item should be removed after pickup")
	}
}

func TestIntegration_PickupMagazine(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	_, magDefID := registerGunAndMag(t, game)
	groundID, err := game.SpawnGroundItem(state.Position{X: 50.5, Y: 50}, magDefID, 1, 8)
	if err != nil {
		t.Fatalf("SpawnGroundItem: %v", err)
	}

	inp := noopInput()
	inp.PickupEntityID = int64(groundID)
	tickWithInput(t, game, pid, inp, 1)

	inv, _ := game.PlayerInventory(pid)
	if inv.Items[0].IsEmpty() {
		t.Fatal("Expected magazine in item slot 0")
	}
	if inv.Items[0].Ammo != 8 {
		t.Errorf("Expected Ammo=8, got %d", inv.Items[0].Ammo)
	}
}

func TestIntegration_PickupOutOfRange(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	gunDefID, _ := registerGunAndMag(t, game)
	// 10 units away (PickupRange = 2.0)
	groundID, err := game.SpawnGroundItem(state.Position{X: 60, Y: 50}, gunDefID, 1, 0)
	if err != nil {
		t.Fatalf("SpawnGroundItem: %v", err)
	}

	inp := noopInput()
	inp.PickupEntityID = int64(groundID)
	tickWithInput(t, game, pid, inp, 1)

	inv, _ := game.PlayerInventory(pid)
	if !inv.Weapons[2].IsEmpty() {
		t.Error("Should not pick up weapon out of range")
	}
	if len(game.GroundItemSnapshots()) != 1 {
		t.Error("Ground item should still exist")
	}
}

func TestIntegration_DropWeapon(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	knifeDefID := registerKnifeDef(t, game)
	inv, _ := game.PlayerInventory(pid)
	inv.Weapons[1] = state.WeaponSlot{ItemID: knifeDefID}
	if err := game.SetPlayerInventory(pid, inv); err != nil {
		t.Fatalf("SetPlayerInventory: %v", err)
	}

	inp := noopInput()
	inp.DropSlotIndex = 1
	tickWithInput(t, game, pid, inp, 1)

	inv, _ = game.PlayerInventory(pid)
	if !inv.Weapons[1].IsEmpty() {
		t.Error("Weapon slot 1 should be empty after drop")
	}
	items := game.GroundItemSnapshots()
	if len(items) != 1 {
		t.Fatalf("Expected 1 ground item, got %d", len(items))
	}
	if items[0].GroundItem.ItemDefID != knifeDefID {
		t.Error("Dropped item should be the knife")
	}
}

func TestIntegration_CannotDropFist(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	inp := noopInput()
	inp.DropSlotIndex = 0
	tickWithInput(t, game, pid, inp, 1)

	inv, _ := game.PlayerInventory(pid)
	if inv.Weapons[0].IsEmpty() {
		t.Error("Fist should not be droppable")
	}
	if len(game.GroundItemSnapshots()) != 0 {
		t.Error("No ground items should be created")
	}
}

// ---------------------------------------------------------------------------
// Weapon Switch
// ---------------------------------------------------------------------------

func TestIntegration_WeaponSwitchCycles(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	knifeDefID := registerKnifeDef(t, game)
	gunDefID, _ := registerGunAndMag(t, game)
	inv, _ := game.PlayerInventory(pid)
	inv.Weapons[1] = state.WeaponSlot{ItemID: knifeDefID}
	inv.Weapons[2] = state.WeaponSlot{ItemID: gunDefID}
	if err := game.SetPlayerInventory(pid, inv); err != nil {
		t.Fatalf("SetPlayerInventory: %v", err)
	}

	// Wait past initial switch cooldown
	tickN(t, game, pid, 16)

	switchInp := noopInput()
	switchInp.SwitchWeapon = true

	// 0 → 1
	tickWithInput(t, game, pid, switchInp, 1)
	inv, _ = game.PlayerInventory(pid)
	if inv.CurrentWeaponIndex != 1 {
		t.Errorf("Switch 1: expected index 1, got %d", inv.CurrentWeaponIndex)
	}

	// Wait for cooldown
	tickN(t, game, pid, 15)

	// 1 → 2
	tickWithInput(t, game, pid, switchInp, 1)
	inv, _ = game.PlayerInventory(pid)
	if inv.CurrentWeaponIndex != 2 {
		t.Errorf("Switch 2: expected index 2, got %d", inv.CurrentWeaponIndex)
	}

	tickN(t, game, pid, 15)

	// 2 → 0
	tickWithInput(t, game, pid, switchInp, 1)
	inv, _ = game.PlayerInventory(pid)
	if inv.CurrentWeaponIndex != 0 {
		t.Errorf("Switch 3: expected index 0, got %d", inv.CurrentWeaponIndex)
	}
}

func TestIntegration_WeaponSwitchCooldown(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	knifeDefID := registerKnifeDef(t, game)
	inv, _ := game.PlayerInventory(pid)
	inv.Weapons[1] = state.WeaponSlot{ItemID: knifeDefID}
	if err := game.SetPlayerInventory(pid, inv); err != nil {
		t.Fatalf("SetPlayerInventory: %v", err)
	}

	// Wait past initial cooldown
	tickN(t, game, pid, 16)

	// First switch: 0 → 1
	switchInp := noopInput()
	switchInp.SwitchWeapon = true
	tickWithInput(t, game, pid, switchInp, 1)
	inv, _ = game.PlayerInventory(pid)
	if inv.CurrentWeaponIndex != 1 {
		t.Fatalf("First switch: expected index 1, got %d", inv.CurrentWeaponIndex)
	}

	// Try to switch at +10 ticks (within 15-tick cooldown)
	tickN(t, game, pid, 9)
	tickWithInput(t, game, pid, switchInp, 1) // 10 ticks since last switch

	inv, _ = game.PlayerInventory(pid)
	if inv.CurrentWeaponIndex != 1 {
		t.Errorf("Cooldown should block switch, expected index 1, got %d", inv.CurrentWeaponIndex)
	}
}

// ---------------------------------------------------------------------------
// Combat / Projectile
// ---------------------------------------------------------------------------

func TestIntegration_ProjectileHitsPlayer(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	pidA, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer A: %v", err)
	}
	pidB, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer B: %v", err)
	}

	noop := noopInput()

	// Move A south so the projectile (fired north) travels toward B's hitbox.
	// B stays at the spawn point (50,50); A moves away and then fires back
	// north toward B's position.
	moveDown := noopInput()
	moveDown.MoveVertical = 1.0 // south in Y-down coords
	for i := 0; i < 60; i++ {
		game.SetPlayerInput(pidA, moveDown)
		game.SetPlayerInput(pidB, noop)
		game.Update(ports.DeltaTime)
	}

	// Verify B's health before fire
	healthBefore, ok := game.PlayerHealth(pidB)
	if !ok {
		t.Fatal("PlayerHealth B not found")
	}
	if healthBefore != 100 {
		t.Fatalf("B health should be 100 before combat, got %d", healthBefore)
	}

	// A fires north (dir=0 default, fist: Damage=5, Speed=50, Range=5).
	// A is now at ~(50,55). Projectile spawns at (50, 54.5), moves north
	// toward B's hitbox at (50, 50). Distance ~4.5, speed ~0.833/tick → hit ~tick 5.
	fireInp := noopInput()
	fireInp.Fire = true
	game.SetPlayerInput(pidA, fireInp)
	game.SetPlayerInput(pidB, noop)
	game.Update(ports.DeltaTime)

	// Wait for projectile to reach B
	for i := 0; i < 10; i++ {
		game.SetPlayerInput(pidA, noop)
		game.SetPlayerInput(pidB, noop)
		game.Update(ports.DeltaTime)
	}

	healthAfter, _ := game.PlayerHealth(pidB)
	if healthAfter >= healthBefore {
		t.Errorf("B should take damage: before=%d after=%d", healthBefore, healthAfter)
	}
	expectedHealth := state.Health(int(healthBefore) - 5) // FistSpec.Damage = 5
	if healthAfter != expectedHealth {
		t.Errorf("Expected health=%d, got %d", expectedHealth, healthAfter)
	}
}

func TestIntegration_ProjectileDoesNotHitOwner(t *testing.T) {
	game := newTestGame(t, nil, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	// Fire fist projectile (dir=0, north)
	fireInp := noopInput()
	fireInp.Fire = true
	tickWithInput(t, game, pid, fireInp, 1)

	// Wait for projectile to travel and expire (FistSpec TTL = 6 ticks)
	tickN(t, game, pid, 10)

	health, _ := game.PlayerHealth(pid)
	if health != 100 {
		t.Errorf("Own health should remain 100, got %d", health)
	}
	if n := game.ProjectileCount(); n != 0 {
		t.Errorf("Projectile should have expired, count=%d", n)
	}
}

func TestIntegration_ProjectileDestroyedByWall(t *testing.T) {
	// Wall directly north of player
	walls := []engine.WallConfig{{
		Center:   vector.Vector2D{X: 50, Y: 45},
		HalfSize: vector.Vector2D{X: 5, Y: 1},
	}}
	game := newTestGame(t, walls, 50, 50)
	pid, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("JoinPlayer: %v", err)
	}

	// Fire north toward wall (default dir=0)
	fireInp := noopInput()
	fireInp.Fire = true
	tickWithInput(t, game, pid, fireInp, 1)

	if n := game.ProjectileCount(); n != 1 {
		t.Fatalf("Expected 1 projectile after fire, got %d", n)
	}

	// Wait for projectile to hit wall
	// Distance to wall: player at Y=50, wall top edge at Y=44.
	// Projectile spawns at Y=49.5, wall bottom edge at Y=46.
	// Distance = 49.5 - 46 = 3.5. Speed = 50/60 per tick ≈ 0.833.
	// Ticks to reach: 3.5/0.833 ≈ 4.2 → hits at tick 5.
	tickN(t, game, pid, 10)

	if n := game.ProjectileCount(); n != 0 {
		t.Errorf("Projectile should be destroyed by wall, count=%d", n)
	}
}
