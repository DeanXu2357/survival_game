package system

import (
	"testing"

	"survival/internal/engine/state"
	"survival/internal/engine/vector"
)

func TestWeaponFire_SpawnsProjectileOnFire(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 5

	playerID, _ := world.CreatePlayer(state.CreatePlayer{
		Position:      state.Position{X: 50, Y: 50},
		Direction:     0,
		MovementSpeed: 5.0,
		RotationSpeed: 2.0,
		Radius:        0.5,
		Health:        100,
	})
	world.ApplyCommands()

	wf := NewWeaponFireSystem(world, &tick)

	world.SetInput(playerID, state.Input{Fire: true})
	world.SyncInputBuffer()

	wf.Update(1.0 / 60.0)
	world.ApplyCommands()

	// Check that a projectile was created
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
	if foundProj.Speed != defaultProjectileSpeed {
		t.Errorf("Expected Speed=%f, got %f", defaultProjectileSpeed, foundProj.Speed)
	}
	if foundProj.Damage != defaultProjectileDamage {
		t.Errorf("Expected Damage=%d, got %d", defaultProjectileDamage, foundProj.Damage)
	}

	dt := 1.0 / 60.0
	expectedExpiredAt := tick + uint64(defaultProjectileRange/defaultProjectileSpeed/dt)
	if foundProj.ExpiredAt != expectedExpiredAt {
		t.Errorf("Expected ExpiredAt=%d, got %d", expectedExpiredAt, foundProj.ExpiredAt)
	}
}

func TestWeaponFire_NoProjectileWithoutFireInput(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 5

	playerID, _ := world.CreatePlayer(state.CreatePlayer{
		Position:      state.Position{X: 50, Y: 50},
		Direction:     0,
		MovementSpeed: 5.0,
		RotationSpeed: 2.0,
		Radius:        0.5,
		Health:        100,
	})
	world.ApplyCommands()

	wf := NewWeaponFireSystem(world, &tick)

	world.SetInput(playerID, state.Input{Fire: false, MoveHorizontal: 1})
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

	playerID, _ := world.CreatePlayer(state.CreatePlayer{
		Position:      playerPos,
		Direction:     playerDir,
		MovementSpeed: 5.0,
		RotationSpeed: 2.0,
		Radius:        0.5,
		Health:        100,
	})
	world.ApplyCommands()

	wf := NewWeaponFireSystem(world, &tick)

	world.SetInput(playerID, state.Input{Fire: true})
	world.SyncInputBuffer()

	wf.Update(1.0 / 60.0)
	world.ApplyCommands()

	// Find the projectile entity
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
