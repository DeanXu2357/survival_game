package system

import (
	"testing"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/vector"
)

func createPlayerWithWeapons(t *testing.T, world *state.World) (playerID state.EntityID, fistDefID state.EntityID, knifeDef state.EntityID, gunDef state.EntityID) {
	t.Helper()

	fistDefID = createFistDef(t, world)
	knifeDef = createWeaponItemDef(t, world, "Knife", state.WeaponConfig{
		Type: state.WeaponTypeKnife, Range: 1, FireRate: 7, Damage: 15, Speed: 30,
	})
	gunDef = createWeaponItemDef(t, world, "Pistol", state.WeaponConfig{
		Type: state.WeaponTypeGun, Range: 20, FireRate: 2, Damage: 30, Speed: 30,
	})

	playerID, ok := world.CreatePlayer(state.CreatePlayer{
		Position:      state.Position{X: 50, Y: 50},
		Direction:     0,
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

	// Equip knife and gun
	inv, _ := world.Inventory.Get(playerID)
	inv.Weapons[1] = state.WeaponSlot{ItemID: knifeDef}
	inv.Weapons[2] = state.WeaponSlot{ItemID: gunDef}
	inv.CurrentWeaponIndex = 2 // default to Gun
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	return playerID, fistDefID, knifeDef, gunDef
}

func createPlayerAtPosDir(t *testing.T, world *state.World, pos state.Position, dir state.Direction) (playerID state.EntityID, fistDefID state.EntityID, knifeDef state.EntityID, gunDef state.EntityID) {
	t.Helper()

	fistDefID = createFistDef(t, world)
	knifeDef = createWeaponItemDef(t, world, "Knife", state.WeaponConfig{
		Type: state.WeaponTypeKnife, Range: 1, FireRate: 7, Damage: 15, Speed: 30,
	})
	gunDef = createWeaponItemDef(t, world, "Pistol", state.WeaponConfig{
		Type: state.WeaponTypeGun, Range: 20, FireRate: 2, Damage: 30, Speed: 30,
	})

	playerID, ok := world.CreatePlayer(state.CreatePlayer{
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

	inv, _ := world.Inventory.Get(playerID)
	inv.Weapons[1] = state.WeaponSlot{ItemID: knifeDef}
	inv.Weapons[2] = state.WeaponSlot{ItemID: gunDef}
	inv.CurrentWeaponIndex = 2
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()

	return playerID, fistDefID, knifeDef, gunDef
}

func switchWeapon(t *testing.T, world *state.World, playerID state.EntityID, weaponIndex int) {
	t.Helper()
	inv, ok := world.Inventory.Get(playerID)
	if !ok {
		t.Fatal("Inventory not found")
	}
	inv.CurrentWeaponIndex = weaponIndex
	world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	world.ApplyCommands()
}

func TestWeaponFire_SpawnsProjectileOnFire(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 5

	playerID, _, _, _ := createPlayerWithWeapons(t, world)

	wf := NewWeaponFireSystem(world, &tick)

	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()

	wf.Update(1.0 / 60.0)
	world.ApplyCommands()

	projectileCount := 0
	var foundProj state.ProjectileData
	for _, proj := range world.Projectile.All() {
		projectileCount++
		foundProj = proj
	}

	if projectileCount != 1 {
		t.Fatalf("Expected 1 projectile, got %d", projectileCount)
	}

	if foundProj.OwnerID != playerID {
		t.Errorf("Expected OwnerID=%d, got %d", playerID, foundProj.OwnerID)
	}

	// Gun spec: Speed=30
	expectedSpeed := 30.0
	expectedDamage := 30
	expectedRange := 20.0

	if foundProj.Speed != expectedSpeed {
		t.Errorf("Expected Speed=%f, got %f", expectedSpeed, foundProj.Speed)
	}
	if foundProj.Damage != expectedDamage {
		t.Errorf("Expected Damage=%d, got %d", expectedDamage, foundProj.Damage)
	}

	expectedExpiredAt := tick + uint64(expectedRange/expectedSpeed*ports.TargetTickRate)
	if foundProj.ExpiredAt != expectedExpiredAt {
		t.Errorf("Expected ExpiredAt=%d, got %d", expectedExpiredAt, foundProj.ExpiredAt)
	}
}

func TestWeaponFire_NoProjectileWithoutFireInput(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 5

	playerID, _, _, _ := createPlayerWithWeapons(t, world)

	wf := NewWeaponFireSystem(world, &tick)

	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: false, MoveHorizontal: 1})
	world.SyncInputBuffer()

	wf.Update(1.0 / 60.0)
	world.ApplyCommands()

	projectileCount := 0
	for range world.Projectile.All() {
		projectileCount++
	}

	if projectileCount != 0 {
		t.Errorf("Expected 0 projectiles, got %d", projectileCount)
	}
}

func TestWeaponFire_ProjectileSpawnsAtPlayerPosition(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 5

	playerPos := state.Position{X: 30, Y: 40}
	playerDir := state.Direction(1.5)

	playerID, _, _, _ := createPlayerAtPosDir(t, world, playerPos, playerDir)

	wf := NewWeaponFireSystem(world, &tick)

	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()

	wf.Update(1.0 / 60.0)
	world.ApplyCommands()

	for entityID, proj := range world.Projectile.All() {
		if proj.OwnerID != playerID {
			continue
		}

		pos, ok := world.Position.Get(entityID)
		if !ok {
			t.Fatal("projectile position not found")
		}
		fwd := vector.Forward(float64(playerDir)).Scale(0.5)
		expectedPos := vector.Vector2D(playerPos).Add(fwd)
		if !floatEquals(pos.X, expectedPos.X, 1e-6) || !floatEquals(pos.Y, expectedPos.Y, 1e-6) {
			t.Errorf("Expected projectile at (%f, %f), got (%f, %f)", expectedPos.X, expectedPos.Y, pos.X, pos.Y)
		}

		dir, ok := world.Direction.Get(entityID)
		if !ok {
			t.Fatal("projectile direction not found")
		}
		if !floatEquals(float64(dir), float64(playerDir), 1e-6) {
			t.Errorf("Expected projectile direction=%f, got %f", float64(playerDir), float64(dir))
		}
		return
	}
	t.Fatal("no projectile found with matching OwnerID")
}

func TestWeaponFire_FireRateLimiting(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 100

	playerID, _, _, _ := createPlayerWithWeapons(t, world)

	wf := NewWeaponFireSystem(world, &tick)

	// First fire should succeed
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()
	wf.Update(1.0 / 60.0)
	world.ApplyCommands()

	projectileCount := 0
	for range world.Projectile.All() {
		projectileCount++
	}
	if projectileCount != 1 {
		t.Fatalf("Expected 1 projectile after first fire, got %d", projectileCount)
	}

	// Second fire on next tick should be blocked by fire rate (Gun fireRate=2, interval=30 ticks)
	tick++
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()
	wf.Update(1.0 / 60.0)
	world.ApplyCommands()

	projectileCount = 0
	for range world.Projectile.All() {
		projectileCount++
	}
	if projectileCount != 1 {
		t.Errorf("Expected still 1 projectile (fire rate blocked), got %d", projectileCount)
	}
}

func TestWeaponFire_PerWeaponLastFireTick(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 100

	playerID, _, _, _ := createPlayerWithWeapons(t, world)

	wf := NewWeaponFireSystem(world, &tick)

	// Fire gun at tick 100
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()
	wf.Update(1.0 / 60.0)
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(playerID)
	if inv.Weapons[2].LastFireTick != 100 {
		t.Errorf("Expected Gun LastFireTick=100, got %d", inv.Weapons[2].LastFireTick)
	}

	// Switch to knife and fire - should work immediately since knife has its own LastFireTick
	switchWeapon(t, world, playerID, 1)
	tick++
	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()
	wf.Update(1.0 / 60.0)
	world.ApplyCommands()

	projectileCount := 0
	for range world.Projectile.All() {
		projectileCount++
	}
	if projectileCount != 2 {
		t.Errorf("Expected 2 projectiles (gun + knife), got %d", projectileCount)
	}

	inv, _ = world.Inventory.Get(playerID)
	if inv.Weapons[1].LastFireTick != 101 {
		t.Errorf("Expected Knife LastFireTick=101, got %d", inv.Weapons[1].LastFireTick)
	}
}

func TestWeaponFire_KnifeSpawnsProjectile(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 5

	playerID, _, _, _ := createPlayerWithWeapons(t, world)
	switchWeapon(t, world, playerID, 1) // Knife

	wf := NewWeaponFireSystem(world, &tick)

	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()
	wf.Update(1.0 / 60.0)
	world.ApplyCommands()

	projectileCount := 0
	var foundProj state.ProjectileData
	for _, proj := range world.Projectile.All() {
		projectileCount++
		foundProj = proj
	}

	if projectileCount != 1 {
		t.Fatalf("Expected 1 projectile for knife, got %d", projectileCount)
	}

	if foundProj.Speed != 30.0 {
		t.Errorf("Expected knife projectile Speed=30, got %f", foundProj.Speed)
	}
	if foundProj.Damage != 15 {
		t.Errorf("Expected knife projectile Damage=15, got %d", foundProj.Damage)
	}
	if foundProj.Range != 1.0 {
		t.Errorf("Expected knife projectile Range=1, got %f", foundProj.Range)
	}
}

func TestWeaponFire_FistSpawnsProjectile(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 5

	playerID, _, _, _ := createPlayerWithWeapons(t, world)
	switchWeapon(t, world, playerID, 0) // Fist

	wf := NewWeaponFireSystem(world, &tick)

	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()
	wf.Update(1.0 / 60.0)
	world.ApplyCommands()

	projectileCount := 0
	var foundProj state.ProjectileData
	for _, proj := range world.Projectile.All() {
		projectileCount++
		foundProj = proj
	}

	if projectileCount != 1 {
		t.Fatalf("Expected 1 projectile for fist, got %d", projectileCount)
	}

	if foundProj.Speed != 50.0 {
		t.Errorf("Expected fist projectile Speed=50, got %f", foundProj.Speed)
	}
	if foundProj.Damage != 5 {
		t.Errorf("Expected fist projectile Damage=5, got %d", foundProj.Damage)
	}
	if foundProj.Range != 5.0 {
		t.Errorf("Expected fist projectile Range=5, got %f", foundProj.Range)
	}

	expectedExpiredAt := tick + uint64(5.0/50.0*ports.TargetTickRate)
	if foundProj.ExpiredAt != expectedExpiredAt {
		t.Errorf("Expected fist ExpiredAt=%d, got %d", expectedExpiredAt, foundProj.ExpiredAt)
	}
}

func TestWeaponFire_FistFallback(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 5

	// Create player with only fist (no other weapons)
	fistDefID := createFistDef(t, world)
	playerID, ok := world.CreatePlayer(state.CreatePlayer{
		Position:      state.Position{X: 50, Y: 50},
		Direction:     0,
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

	wf := NewWeaponFireSystem(world, &tick)

	world.SetInput(playerID, state.Input{PickupEntityID: state.NoPickup, DropSlotIndex: state.NoDrop, Fire: true})
	world.SyncInputBuffer()
	wf.Update(1.0 / 60.0)
	world.ApplyCommands()

	projectileCount := 0
	var foundProj state.ProjectileData
	for _, proj := range world.Projectile.All() {
		projectileCount++
		foundProj = proj
	}

	if projectileCount != 1 {
		t.Fatalf("Expected 1 projectile for fist fallback, got %d", projectileCount)
	}

	// Should use FistSpec stats
	if foundProj.Speed != state.FistSpec.Speed {
		t.Errorf("Expected fist Speed=%f, got %f", state.FistSpec.Speed, foundProj.Speed)
	}
	if foundProj.Damage != state.FistSpec.Damage {
		t.Errorf("Expected fist Damage=%d, got %d", state.FistSpec.Damage, foundProj.Damage)
	}
}
