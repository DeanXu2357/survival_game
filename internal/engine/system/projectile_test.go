package system

import (
	"math"
	"testing"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/vector"
)

func setupProjectileWorld() *state.World {
	gridCellSize := 5.0
	gridWidth := 20
	gridHeight := 20
	world := state.NewWorld(gridCellSize, gridWidth, gridHeight)
	world.Width = 100
	world.Height = 100
	// Reserve entity 0 so no real entity gets EntityID(0),
	// which would be confused with "empty" in IsEmpty() checks.
	world.Entity.Alloc()
	return world
}

func spawnProjectile(t *testing.T, world *state.World, pos state.Position, dir state.Direction, proj state.ProjectileData) state.EntityID {
	t.Helper()
	id, ok := world.CreateProjectileEntity(state.CreateProjectile{
		Position:       pos,
		Direction:      dir,
		ProjectileData: proj,
	})
	if !ok {
		t.Fatal("failed to create projectile entity")
	}
	world.ApplyCommands()
	return id
}

func TestProjectile_MovesCorrectly(t *testing.T) {
	world := setupProjectileWorld()
	var tick ports.Tick = 1
	ps := NewProjectileSystem(world, &tick)

	// Direction 0 means facing "north" (sin(0)=0, -cos(0)=-1), so Y decreases
	projID := spawnProjectile(t, world, state.Position{X: 50, Y: 50}, 0, state.ProjectileData{
		Speed:     20.0,
		Range:     50.0,
		Damage:    25,
		OwnerID:   0,
		Height:    1.5,
		ExpiredAt: 1000,
	})

	dt := 1.0 / 60.0
	ps.Update(dt)
	world.ApplyCommands()

	pos, ok := world.Position.Get(projID)
	if !ok {
		t.Fatal("projectile position not found after update")
	}

	expectedY := 50.0 + (-1.0)*20.0*dt // cos(0)=1, so Y -= speed*dt
	if !floatEquals(pos.X, 50, 1e-6) {
		t.Errorf("Expected X=50, got %f", pos.X)
	}
	if !floatEquals(pos.Y, expectedY, 1e-6) {
		t.Errorf("Expected Y=%f, got %f", expectedY, pos.Y)
	}
}

func TestProjectile_MovesMultipleTicks(t *testing.T) {
	world := setupProjectileWorld()
	var tick ports.Tick = 1
	ps := NewProjectileSystem(world, &tick)

	projID := spawnProjectile(t, world, state.Position{X: 50, Y: 50}, 0, state.ProjectileData{
		Speed:     20.0,
		Range:     50.0,
		Damage:    25,
		OwnerID:   0,
		Height:    1.5,
		ExpiredAt: 1000,
	})

	dt := 1.0 / 60.0
	ticks := 10
	for i := 0; i < ticks; i++ {
		ps.Update(dt)
		world.ApplyCommands()
	}

	pos, ok := world.Position.Get(projID)
	if !ok {
		t.Fatal("projectile position not found")
	}

	expectedY := 50.0 - 20.0*dt*float64(ticks)
	if !floatEquals(pos.Y, expectedY, 1e-4) {
		t.Errorf("Expected Y=%f after %d ticks, got %f", expectedY, ticks, pos.Y)
	}
}

func TestProjectile_DestroyedOnTTLExpiry(t *testing.T) {
	world := setupProjectileWorld()
	var tick ports.Tick = 10
	ps := NewProjectileSystem(world, &tick)

	projID := spawnProjectile(t, world, state.Position{X: 50, Y: 50}, 0, state.ProjectileData{
		Speed:     20.0,
		Range:     50.0,
		Damage:    25,
		OwnerID:   0,
		Height:    1.5,
		ExpiredAt: 10, // expires at tick 10
	})

	ps.Update(1.0 / 60.0)
	world.ApplyCommands()

	// Projectile should be destroyed
	_, ok := world.Projectile.Get(projID)
	if ok {
		t.Error("projectile should be destroyed after TTL expiry")
	}
	if world.Entity.IsAlive(projID) {
		t.Error("projectile entity should not be alive after TTL expiry")
	}
}

func TestProjectile_DestroyedOnWallCollision(t *testing.T) {
	world := setupProjectileWorld()
	var tick ports.Tick = 1
	ps := NewProjectileSystem(world, &tick)

	// Add wall at Y=45 (center=45, halfSize=2x2, so Y range is 43-47)
	addWall(world, 50, 45, 2, 2)

	// Projectile at Y=47.5 moving north (dir=0) at speed 20
	// After 1 tick: Y = 47.5 - 20/60 = 47.5 - 0.333 = 47.167 (inside wall 43-47)
	projID := spawnProjectile(t, world, state.Position{X: 50, Y: 47.5}, 0, state.ProjectileData{
		Speed:     20.0,
		Range:     50.0,
		Damage:    25,
		OwnerID:   0,
		Height:    1.5,
		ExpiredAt: 1000,
	})

	// Run enough ticks for projectile to reach the wall
	dt := 1.0 / 60.0
	for i := 0; i < 5; i++ {
		ps.Update(dt)
		world.ApplyCommands()
	}

	_, ok := world.Projectile.Get(projID)
	if ok {
		t.Error("projectile should be destroyed after hitting a wall")
	}
}

func TestProjectile_HitsPlayer_NotOwner(t *testing.T) {
	world := setupProjectileWorld()
	var tick ports.Tick = 1
	ps := NewProjectileSystem(world, &tick)

	// Create owner player
	ownerID, _ := world.CreatePlayer(state.CreatePlayer{
		Position:      state.Position{X: 10, Y: 50},
		Direction:     0,
		MovementSpeed: 5.0,
		RotationSpeed: 2.0,
		Radius:        0.5,
		Health:        100,
	})
	world.ApplyCommands()

	// Create target player at Y=48 (center)
	targetID, _ := world.CreatePlayer(state.CreatePlayer{
		Position:      state.Position{X: 50, Y: 48},
		Direction:     0,
		MovementSpeed: 5.0,
		RotationSpeed: 2.0,
		Radius:        0.5,
		Health:        100,
	})
	world.ApplyCommands()

	// Projectile moving north toward target
	projID := spawnProjectile(t, world, state.Position{X: 50, Y: 50}, 0, state.ProjectileData{
		Speed:     20.0,
		Range:     50.0,
		Damage:    25,
		OwnerID:   ownerID,
		Height:    1.5,
		ExpiredAt: 1000,
	})

	dt := 1.0 / 60.0
	for i := 0; i < 30; i++ {
		ps.Update(dt)
		world.ApplyCommands()
	}

	// Projectile should be destroyed
	_, projExists := world.Projectile.Get(projID)
	if projExists {
		t.Error("projectile should be destroyed after hitting target player")
	}

	// Target health should be reduced
	targetHealth, ok := world.Health.Get(targetID)
	if !ok {
		t.Fatal("target health not found")
	}
	if int(targetHealth) != 75 {
		t.Errorf("Expected target health=75, got %d", int(targetHealth))
	}

	// Owner health should be unchanged
	ownerHealth, ok := world.Health.Get(ownerID)
	if !ok {
		t.Fatal("owner health not found")
	}
	if int(ownerHealth) != 100 {
		t.Errorf("Expected owner health=100, got %d", int(ownerHealth))
	}
}

func TestProjectile_DoesNotHitOwner(t *testing.T) {
	world := setupProjectileWorld()
	var tick ports.Tick = 1
	ps := NewProjectileSystem(world, &tick)

	// Create owner at the projectile spawn point
	ownerID, _ := world.CreatePlayer(state.CreatePlayer{
		Position:      state.Position{X: 50, Y: 50},
		Direction:     0,
		MovementSpeed: 5.0,
		RotationSpeed: 2.0,
		Radius:        0.5,
		Health:        100,
	})
	world.ApplyCommands()

	// Projectile at same position as owner
	spawnProjectile(t, world, state.Position{X: 50, Y: 50}, 0, state.ProjectileData{
		Speed:     20.0,
		Range:     50.0,
		Damage:    25,
		OwnerID:   ownerID,
		Height:    1.5,
		ExpiredAt: 1000,
	})

	ps.Update(1.0 / 60.0)
	world.ApplyCommands()

	// Owner health should be unchanged
	ownerHealth, ok := world.Health.Get(ownerID)
	if !ok {
		t.Fatal("owner health not found")
	}
	if int(ownerHealth) != 100 {
		t.Errorf("Expected owner health=100, got %d", int(ownerHealth))
	}
}

func TestProjectile_DirectionAffectsMovement(t *testing.T) {
	world := setupProjectileWorld()
	var tick ports.Tick = 1
	ps := NewProjectileSystem(world, &tick)

	// Direction Pi/2 means facing "east" (sin(Pi/2)=1, -cos(Pi/2)=0)
	dir := state.Direction(math.Pi / 2)
	projID := spawnProjectile(t, world, state.Position{X: 50, Y: 50}, dir, state.ProjectileData{
		Speed:     20.0,
		Range:     50.0,
		Damage:    25,
		OwnerID:   0,
		Height:    1.5,
		ExpiredAt: 1000,
	})

	dt := 1.0 / 60.0
	ps.Update(dt)
	world.ApplyCommands()

	pos, ok := world.Position.Get(projID)
	if !ok {
		t.Fatal("projectile position not found")
	}

	expectedX := 50.0 + math.Sin(math.Pi/2)*20.0*dt
	expectedY := 50.0 + (-math.Cos(math.Pi/2))*20.0*dt

	if !floatEquals(pos.X, expectedX, 1e-6) {
		t.Errorf("Expected X=%f, got %f", expectedX, pos.X)
	}
	if !floatEquals(pos.Y, expectedY, 1e-6) {
		t.Errorf("Expected Y=%f, got %f", expectedY, pos.Y)
	}
}

func TestProjectile_WallCollisionAtBoundary(t *testing.T) {
	world := setupProjectileWorld()
	var tick ports.Tick = 1
	ps := NewProjectileSystem(world, &tick)

	// Place a large wall covering center of map
	wallID, _ := world.Entity.Alloc()
	collider := state.Collider{
		Center:    state.Position{X: 50, Y: 30},
		HalfSize:  vector.Vector2D{X: 10, Y: 5},
		ShapeType: state.ColliderBox,
	}
	world.Collider.Upsert(wallID, collider)
	world.EntityMeta.Upsert(wallID, state.WallMeta)
	min, max := collider.BoundingBox()
	world.Grid.Add(wallID, state.Bounds{
		MinX: min.X, MinY: min.Y,
		MaxX: max.X, MaxY: max.Y,
	}, state.LayerStatic)

	// Projectile heading straight into the wall
	projID := spawnProjectile(t, world, state.Position{X: 50, Y: 36}, 0, state.ProjectileData{
		Speed:     20.0,
		Range:     50.0,
		Damage:    25,
		OwnerID:   0,
		Height:    1.5,
		ExpiredAt: 1000,
	})

	dt := 1.0 / 60.0
	for i := 0; i < 10; i++ {
		ps.Update(dt)
		world.ApplyCommands()
	}

	_, ok := world.Projectile.Get(projID)
	if ok {
		pos, _ := world.Position.Get(projID)
		t.Errorf("projectile should be destroyed after hitting wall, pos=(%f, %f)", pos.X, pos.Y)
	}
}
