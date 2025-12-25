package domain

import (
	"math"
	"testing"

	"survival/internal/core/domain/vector"
	"survival/internal/core/ports"
)

func createTestWorld() *World {
	return &World{
		Entity:        NewEntityManager(),
		Position:      *NewComponentManager[Position](),
		Direction:     *NewComponentManager[Direction](),
		MovementSpeed: *NewComponentManager[MovementSpeed](),
		RotationSpeed: *NewComponentManager[RotationSpeed](),
		PlayerShape:   *NewComponentManager[PlayerShape](),
		WallShape:     *NewComponentManager[WallShape](),
		Grid: Grid{
			cellSize:  100,
			width:     10,
			height:    10,
			cellSlice: make([]GridCell, 100),
		},
	}
}

func createPlayerEntity(world *World, pos Position, dir Direction, moveSpeed MovementSpeed, rotSpeed RotationSpeed) EntityID {
	entityID, _ := world.Entity.Alloc()
	world.Position.Add(entityID, pos)
	world.Direction.Add(entityID, dir)
	world.MovementSpeed.Add(entityID, moveSpeed)
	world.RotationSpeed.Add(entityID, rotSpeed)
	world.PlayerShape.Add(entityID, PlayerShape{Radius: 10})
	return entityID
}

const tolerance = 1e-9

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestMovementSystem_Update_MoveUp(t *testing.T) {
	world := createTestWorld()
	buf := NewCommandBuffer()
	ms := NewMovementSystem()

	entityID := createPlayerEntity(world, Position{X: 100, Y: 100}, Direction(0), MovementSpeed(100), RotationSpeed(1))

	input := ports.PlayerInput{MoveUp: true}
	playerInputs := map[EntityID]ports.PlayerInput{entityID: input}

	dt := 1.0
	result := ms.Update(dt, world, buf, playerInputs)

	newPos := result[entityID]
	// Screen coordinates: MoveUp decreases Y
	if newPos.Y >= 100 {
		t.Errorf("MoveUp should decrease Y, got Y=%f, expected < 100", newPos.Y)
	}
	if !floatEquals(newPos.X, 100) {
		t.Errorf("MoveUp should not change X, got X=%f, expected 100", newPos.X)
	}
}

func TestMovementSystem_Update_MoveDown(t *testing.T) {
	world := createTestWorld()
	buf := NewCommandBuffer()
	ms := NewMovementSystem()

	entityID := createPlayerEntity(world, Position{X: 100, Y: 100}, Direction(0), MovementSpeed(100), RotationSpeed(1))

	input := ports.PlayerInput{MoveDown: true}
	playerInputs := map[EntityID]ports.PlayerInput{entityID: input}

	dt := 1.0
	result := ms.Update(dt, world, buf, playerInputs)

	newPos := result[entityID]
	// Screen coordinates: MoveDown increases Y
	if newPos.Y <= 100 {
		t.Errorf("MoveDown should increase Y, got Y=%f, expected > 100", newPos.Y)
	}
}

func TestMovementSystem_Update_MoveLeft(t *testing.T) {
	world := createTestWorld()
	buf := NewCommandBuffer()
	ms := NewMovementSystem()

	entityID := createPlayerEntity(world, Position{X: 100, Y: 100}, Direction(0), MovementSpeed(100), RotationSpeed(1))

	input := ports.PlayerInput{MoveLeft: true}
	playerInputs := map[EntityID]ports.PlayerInput{entityID: input}

	dt := 1.0
	result := ms.Update(dt, world, buf, playerInputs)

	newPos := result[entityID]
	if newPos.X >= 100 {
		t.Errorf("MoveLeft should decrease X, got X=%f, expected < 100", newPos.X)
	}
}

func TestMovementSystem_Update_MoveRight(t *testing.T) {
	world := createTestWorld()
	buf := NewCommandBuffer()
	ms := NewMovementSystem()

	entityID := createPlayerEntity(world, Position{X: 100, Y: 100}, Direction(0), MovementSpeed(100), RotationSpeed(1))

	input := ports.PlayerInput{MoveRight: true}
	playerInputs := map[EntityID]ports.PlayerInput{entityID: input}

	dt := 1.0
	result := ms.Update(dt, world, buf, playerInputs)

	newPos := result[entityID]
	if newPos.X <= 100 {
		t.Errorf("MoveRight should increase X, got X=%f, expected > 100", newPos.X)
	}
}

func TestMovementSystem_Update_RotateLeft(t *testing.T) {
	world := createTestWorld()
	buf := NewCommandBuffer()
	ms := NewMovementSystem()

	initialDir := Direction(0)
	entityID := createPlayerEntity(world, Position{X: 100, Y: 100}, initialDir, MovementSpeed(100), RotationSpeed(1))

	input := ports.PlayerInput{RotateLeft: true}
	playerInputs := map[EntityID]ports.PlayerInput{entityID: input}

	dt := 1.0
	ms.Update(dt, world, buf, playerInputs)

	cmd, ok := buf.Pop()
	if !ok {
		t.Fatal("Expected command in buffer")
	}
	if cmd.Direction == nil {
		t.Fatal("Expected direction in command")
	}
	// RotateLeft increases angle
	if *cmd.Direction <= initialDir {
		t.Errorf("RotateLeft should increase direction, got %f, expected > %f", *cmd.Direction, float64(initialDir))
	}
}

func TestMovementSystem_Update_RotateRight(t *testing.T) {
	world := createTestWorld()
	buf := NewCommandBuffer()
	ms := NewMovementSystem()

	initialDir := Direction(0)
	entityID := createPlayerEntity(world, Position{X: 100, Y: 100}, initialDir, MovementSpeed(100), RotationSpeed(1))

	input := ports.PlayerInput{RotateRight: true}
	playerInputs := map[EntityID]ports.PlayerInput{entityID: input}

	dt := 1.0
	ms.Update(dt, world, buf, playerInputs)

	cmd, ok := buf.Pop()
	if !ok {
		t.Fatal("Expected command in buffer")
	}
	if cmd.Direction == nil {
		t.Fatal("Expected direction in command")
	}
	// RotateRight decreases angle
	if *cmd.Direction >= initialDir {
		t.Errorf("RotateRight should decrease direction, got %f, expected < %f", *cmd.Direction, float64(initialDir))
	}
}

func TestMovementSystem_Update_DiagonalMovement(t *testing.T) {
	world := createTestWorld()
	buf := NewCommandBuffer()
	ms := NewMovementSystem()

	speed := MovementSpeed(100)
	entityID := createPlayerEntity(world, Position{X: 100, Y: 100}, Direction(0), speed, RotationSpeed(1))

	input := ports.PlayerInput{MoveUp: true, MoveRight: true}
	playerInputs := map[EntityID]ports.PlayerInput{entityID: input}

	dt := 1.0
	result := ms.Update(dt, world, buf, playerInputs)

	newPos := result[entityID]

	// Diagonal movement should be normalized (not faster than cardinal)
	dx := newPos.X - 100
	dy := newPos.Y - 100
	distance := math.Sqrt(dx*dx + dy*dy)

	expectedDistance := float64(speed) * dt
	if !floatEquals(distance, expectedDistance) {
		t.Errorf("Diagonal movement distance = %f, expected %f (normalized)", distance, expectedDistance)
	}
}

func TestMovementSystem_Update_NoInput(t *testing.T) {
	world := createTestWorld()
	buf := NewCommandBuffer()
	ms := NewMovementSystem()

	entityID := createPlayerEntity(world, Position{X: 100, Y: 100}, Direction(0), MovementSpeed(100), RotationSpeed(1))

	input := ports.PlayerInput{}
	playerInputs := map[EntityID]ports.PlayerInput{entityID: input}

	dt := 1.0
	result := ms.Update(dt, world, buf, playerInputs)

	newPos := result[entityID]
	if !floatEquals(newPos.X, 100) || !floatEquals(newPos.Y, 100) {
		t.Errorf("No input should not change position, got (%f, %f), expected (100, 100)", newPos.X, newPos.Y)
	}
}

func TestMovementSystem_Update_MissingComponents(t *testing.T) {
	world := createTestWorld()
	buf := NewCommandBuffer()
	ms := NewMovementSystem()

	// Create entity without all required components
	entityID, _ := world.Entity.Alloc()
	world.Position.Add(entityID, Position{X: 100, Y: 100})
	// Missing Direction, MovementSpeed, RotationSpeed

	input := ports.PlayerInput{MoveUp: true}
	playerInputs := map[EntityID]ports.PlayerInput{entityID: input}

	dt := 1.0
	result := ms.Update(dt, world, buf, playerInputs)

	// Should skip entity with missing components
	if _, exists := result[entityID]; exists {
		t.Error("Entity with missing components should be skipped")
	}
	if buf.Len() != 0 {
		t.Error("No commands should be pushed for entity with missing components")
	}
}

func TestMovementSystem_Update_CommandBufferPush(t *testing.T) {
	world := createTestWorld()
	buf := NewCommandBuffer()
	ms := NewMovementSystem()

	entityID := createPlayerEntity(world, Position{X: 100, Y: 100}, Direction(0), MovementSpeed(100), RotationSpeed(1))

	input := ports.PlayerInput{MoveUp: true, RotateLeft: true}
	playerInputs := map[EntityID]ports.PlayerInput{entityID: input}

	dt := 1.0
	ms.Update(dt, world, buf, playerInputs)

	if buf.Len() != 1 {
		t.Errorf("Expected 1 command in buffer, got %d", buf.Len())
	}

	cmd, ok := buf.Pop()
	if !ok {
		t.Fatal("Expected command in buffer")
	}
	if cmd.Type != UpdatePlayerCommand {
		t.Errorf("Expected UpdatePlayerCommand, got %v", cmd.Type)
	}
	if cmd.EntityID != entityID {
		t.Errorf("Expected EntityID %v, got %v", entityID, cmd.EntityID)
	}
	if cmd.Position == nil {
		t.Error("Expected Position in command")
	}
	if cmd.Direction == nil {
		t.Error("Expected Direction in command")
	}
}

func TestCircleAABBCollision_NoCollision(t *testing.T) {
	center := Position{X: 0, Y: 0}
	radius := 10.0
	wallMin := vector.Vector2D{X: 50, Y: 50}
	wallMax := vector.Vector2D{X: 100, Y: 100}

	collides, pushOut := circleAABBCollision(center, radius, wallMin, wallMax)

	if collides {
		t.Error("Should not collide when circle is far from AABB")
	}
	if pushOut.X != 0 || pushOut.Y != 0 {
		t.Errorf("PushOut should be zero, got (%f, %f)", pushOut.X, pushOut.Y)
	}
}

func TestCircleAABBCollision_EdgeCollision(t *testing.T) {
	center := Position{X: 45, Y: 75}
	radius := 10.0
	wallMin := vector.Vector2D{X: 50, Y: 50}
	wallMax := vector.Vector2D{X: 100, Y: 100}

	collides, pushOut := circleAABBCollision(center, radius, wallMin, wallMax)

	if !collides {
		t.Error("Should collide when circle overlaps AABB edge")
	}
	// Push should be to the left (negative X direction)
	if pushOut.X >= 0 {
		t.Errorf("PushOut.X should be negative, got %f", pushOut.X)
	}
}

func TestCircleAABBCollision_Penetration(t *testing.T) {
	center := Position{X: 55, Y: 75}
	radius := 10.0
	wallMin := vector.Vector2D{X: 50, Y: 50}
	wallMax := vector.Vector2D{X: 100, Y: 100}

	collides, pushOut := circleAABBCollision(center, radius, wallMin, wallMax)

	if !collides {
		t.Error("Should collide when circle penetrates AABB")
	}

	// After applying pushOut, circle should no longer collide
	newCenter := Position{
		X: center.X + pushOut.X,
		Y: center.Y + pushOut.Y,
	}
	collides2, _ := circleAABBCollision(newCenter, radius, wallMin, wallMax)
	if collides2 {
		t.Error("After applying pushOut, circle should no longer collide")
	}
}

func TestMovementSystem_Update_WallCollision(t *testing.T) {
	world := createTestWorld()
	buf := NewCommandBuffer()
	ms := NewMovementSystem()

	// Create player that will just touch the wall edge after moving
	playerRadius := 10.0
	// Player at X=85, moves right by 10 (speed=10, dt=1), ends at X=95
	// Wall left edge is at X=100, so player edge at X=95+10=105 would penetrate
	// Collision should push player back to X=100-10=90
	entityID := createPlayerEntity(world, Position{X: 85, Y: 100}, Direction(0), MovementSpeed(10), RotationSpeed(1))

	// Create a wall entity
	wallEntityID, _ := world.Entity.Alloc()
	world.WallShape.Add(wallEntityID, WallShape{
		Center:   Position{X: 150, Y: 100},
		HalfSize: vector.Vector2D{X: 50, Y: 50},
	})

	// Add wall to grid - wall spans X: 100-200, Y: 50-150
	world.Grid.Add(wallEntityID, Bounds{MinX: 100, MinY: 50, MaxX: 200, MaxY: 150}, LayerStatic)

	// Try to move right into the wall
	input := ports.PlayerInput{MoveRight: true}
	playerInputs := map[EntityID]ports.PlayerInput{entityID: input}

	dt := 1.0
	result := ms.Update(dt, world, buf, playerInputs)

	newPos := result[entityID]
	// After collision resolution, player center should be at wall edge minus radius
	// Wall left edge is at X=100, so player center should be at X=100-radius=90
	expectedMaxX := 100 - playerRadius
	if newPos.X > expectedMaxX+tolerance {
		t.Errorf("Player should be blocked by wall, got X=%f, expected <= %f", newPos.X, expectedMaxX)
	}
}
