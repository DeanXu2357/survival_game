package system

import (
	"io"
	"log"
	"math"
	"os"
	"testing"

	"survival/internal/engine/state"
	"survival/internal/engine/vector"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func setupTestWorld(playerPos state.Position, playerDir state.Direction) (*state.World, state.EntityID) {
	gridCellSize := 5.0
	gridWidth := 20
	gridHeight := 20
	world := state.NewWorld(gridCellSize, gridWidth, gridHeight)
	world.Width = 100
	world.Height = 100

	playerID, _ := world.CreatePlayer(state.CreatePlayer{
		Position:      playerPos,
		Direction:     playerDir,
		MovementSpeed: 5.0,
		RotationSpeed: 2.0,
		Radius:        0.5,
		Health:        100,
	})
	world.ApplyCommands()
	return world, playerID
}

func addWall(world *state.World, centerX, centerY, halfW, halfH float64) state.EntityID {
	wallID, _ := world.Entity.Alloc()
	collider := state.Collider{
		Center:    state.Position{X: centerX, Y: centerY},
		HalfSize:  vector.Vector2D{X: halfW, Y: halfH},
		ShapeType: state.ColliderBox,
	}
	world.Collider.Upsert(wallID, collider)
	world.EntityMeta.Upsert(wallID, state.WallMeta)

	min, max := collider.BoundingBox()
	world.Grid.Add(wallID, state.Bounds{
		MinX: min.X, MinY: min.Y,
		MaxX: max.X, MaxY: max.Y,
	}, state.LayerStatic)
	return wallID
}

func floatEquals(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestNoInput_PlayerStaysStill(t *testing.T) {
	world, playerID := setupTestWorld(state.Position{X: 50, Y: 50}, 0)
	ms := NewBasicMovementSystem()

	inputs := map[state.EntityID]PlayerInput{
		playerID: {},
	}

	ms.Update(1.0/60.0, world, inputs)
	world.ApplyCommands()

	pos, _ := world.Position.Get(playerID)
	if !floatEquals(pos.X, 50, 1e-6) || !floatEquals(pos.Y, 50, 1e-6) {
		t.Errorf("Expected position (50, 50), got (%f, %f)", pos.X, pos.Y)
	}
}

func TestVerticalMovement_BlockedByWall(t *testing.T) {
	world, playerID := setupTestWorld(state.Position{X: 50, Y: 50}, 0)
	addWall(world, 50, 45, 2, 1)
	ms := NewBasicMovementSystem()

	inputs := map[state.EntityID]PlayerInput{
		playerID: {MoveVertical: -1, MovementType: MovementTypeAbsolute},
	}

	for i := 0; i < 60; i++ {
		ms.Update(1.0/60.0, world, inputs)
		world.ApplyCommands()
	}

	pos, _ := world.Position.Get(playerID)
	playerRadius := 0.5
	wallBottom := 45 + 1
	expectedMinY := float64(wallBottom) + playerRadius

	if pos.Y < expectedMinY-1e-6 {
		t.Errorf("Player passed through wall: Y=%f, expected >= %f", pos.Y, expectedMinY)
	}
}

func TestHorizontalMovement_BlockedByWall(t *testing.T) {
	world, playerID := setupTestWorld(state.Position{X: 50, Y: 50}, 0)
	addWall(world, 55, 50, 1, 2)
	ms := NewBasicMovementSystem()

	inputs := map[state.EntityID]PlayerInput{
		playerID: {MoveHorizontal: 1, MovementType: MovementTypeAbsolute},
	}

	for i := 0; i < 60; i++ {
		ms.Update(1.0/60.0, world, inputs)
		world.ApplyCommands()
	}

	pos, _ := world.Position.Get(playerID)
	playerRadius := 0.5
	wallLeft := 55 - 1
	expectedMaxX := float64(wallLeft) - playerRadius

	if pos.X > expectedMaxX+1e-6 {
		t.Errorf("Player passed through wall: X=%f, expected <= %f", pos.X, expectedMaxX)
	}
}

func TestFreeMovement_NoWalls(t *testing.T) {
	world, playerID := setupTestWorld(state.Position{X: 50, Y: 50}, 0)
	ms := NewBasicMovementSystem()

	dt := 1.0 / 60.0
	speed := 5.0
	expectedDelta := speed * dt

	inputs := map[state.EntityID]PlayerInput{
		playerID: {MoveHorizontal: 1, MovementType: MovementTypeAbsolute},
	}

	ms.Update(dt, world, inputs)
	world.ApplyCommands()

	pos, _ := world.Position.Get(playerID)
	if !floatEquals(pos.X, 50+expectedDelta, 1e-6) {
		t.Errorf("Expected X=%f, got %f", 50+expectedDelta, pos.X)
	}
	if !floatEquals(pos.Y, 50, 1e-6) {
		t.Errorf("Expected Y=50, got %f", pos.Y)
	}
}

func TestDiagonalMovement_PartialBlock(t *testing.T) {
	world, playerID := setupTestWorld(state.Position{X: 50, Y: 50}, 0)
	addWall(world, 55, 50, 1, 5)
	ms := NewBasicMovementSystem()

	inputs := map[state.EntityID]PlayerInput{
		playerID: {MoveHorizontal: 1, MoveVertical: 1, MovementType: MovementTypeAbsolute},
	}

	for i := 0; i < 60; i++ {
		ms.Update(1.0/60.0, world, inputs)
		world.ApplyCommands()
	}

	pos, _ := world.Position.Get(playerID)
	playerRadius := 0.5
	wallLeft := 55 - 1
	expectedMaxX := float64(wallLeft) - playerRadius

	if pos.X > expectedMaxX+1e-6 {
		t.Errorf("Player X passed through wall: X=%f, expected <= %f", pos.X, expectedMaxX)
	}
	if pos.Y <= 50 {
		t.Errorf("Player should have moved down (positive Y), got Y=%f", pos.Y)
	}
}

func TestRotation_NoMovement(t *testing.T) {
	world, playerID := setupTestWorld(state.Position{X: 50, Y: 50}, 0)
	ms := NewBasicMovementSystem()

	dt := 1.0 / 60.0
	rotSpeed := 2.0
	expectedRotation := rotSpeed * dt

	inputs := map[state.EntityID]PlayerInput{
		playerID: {LookHorizontal: 1},
	}

	ms.Update(dt, world, inputs)
	world.ApplyCommands()

	pos, _ := world.Position.Get(playerID)
	dir, _ := world.Direction.Get(playerID)

	if !floatEquals(pos.X, 50, 1e-6) || !floatEquals(pos.Y, 50, 1e-6) {
		t.Errorf("Position should be unchanged, got (%f, %f)", pos.X, pos.Y)
	}
	if !floatEquals(float64(dir), expectedRotation, 1e-6) {
		t.Errorf("Expected direction=%f, got %f", expectedRotation, float64(dir))
	}
}

func TestRelativeMovement_ForwardMeansPlayerDirection(t *testing.T) {
	direction := math.Pi / 2
	world, playerID := setupTestWorld(state.Position{X: 50, Y: 50}, state.Direction(direction))
	ms := NewBasicMovementSystem()

	dt := 1.0 / 60.0
	speed := 5.0
	expectedDelta := speed * dt

	inputs := map[state.EntityID]PlayerInput{
		playerID: {MoveVertical: 1, MovementType: MovementTypeRelative},
	}

	ms.Update(dt, world, inputs)
	world.ApplyCommands()

	pos, _ := world.Position.Get(playerID)

	fwdX := math.Sin(direction)
	fwdY := -math.Cos(direction)
	expectedX := 50 + fwdX*expectedDelta
	expectedY := 50 + fwdY*expectedDelta

	if !floatEquals(pos.X, expectedX, 1e-6) {
		t.Errorf("Expected X=%f, got %f", expectedX, pos.X)
	}
	if !floatEquals(pos.Y, expectedY, 1e-6) {
		t.Errorf("Expected Y=%f, got %f", expectedY, pos.Y)
	}
}

func TestMovementSpeed_AffectsDistance(t *testing.T) {
	gridCellSize := 5.0
	gridWidth := 20
	gridHeight := 20
	world := state.NewWorld(gridCellSize, gridWidth, gridHeight)
	world.Width = 100
	world.Height = 100

	slowPlayerID, _ := world.CreatePlayer(state.CreatePlayer{
		Position:      state.Position{X: 30, Y: 50},
		Direction:     0,
		MovementSpeed: 2.0,
		RotationSpeed: 2.0,
		Radius:        0.5,
		Health:        100,
	})

	fastPlayerID, _ := world.CreatePlayer(state.CreatePlayer{
		Position:      state.Position{X: 70, Y: 50},
		Direction:     0,
		MovementSpeed: 10.0,
		RotationSpeed: 2.0,
		Radius:        0.5,
		Health:        100,
	})
	world.ApplyCommands()

	ms := NewBasicMovementSystem()
	dt := 1.0 / 60.0

	inputs := map[state.EntityID]PlayerInput{
		slowPlayerID: {MoveHorizontal: 1, MovementType: MovementTypeAbsolute},
		fastPlayerID: {MoveHorizontal: 1, MovementType: MovementTypeAbsolute},
	}

	ms.Update(dt, world, inputs)
	world.ApplyCommands()

	slowPos, _ := world.Position.Get(slowPlayerID)
	fastPos, _ := world.Position.Get(fastPlayerID)

	slowDelta := slowPos.X - 30
	fastDelta := fastPos.X - 70

	if fastDelta <= slowDelta {
		t.Errorf("Fast player should move further: slowDelta=%f, fastDelta=%f", slowDelta, fastDelta)
	}

	expectedSlowDelta := 2.0 * dt
	expectedFastDelta := 10.0 * dt
	if !floatEquals(slowDelta, expectedSlowDelta, 1e-6) {
		t.Errorf("Slow player: expected delta=%f, got %f", expectedSlowDelta, slowDelta)
	}
	if !floatEquals(fastDelta, expectedFastDelta, 1e-6) {
		t.Errorf("Fast player: expected delta=%f, got %f", expectedFastDelta, fastDelta)
	}
}

func TestPlayerInsideWall_PushedOut(t *testing.T) {
	world, playerID := setupTestWorld(state.Position{X: 50, Y: 50}, 0)
	addWall(world, 50, 50, 2, 2)
	ms := NewBasicMovementSystem()

	inputs := map[state.EntityID]PlayerInput{
		playerID: {},
	}

	ms.Update(1.0/60.0, world, inputs)
	world.ApplyCommands()

	pos, _ := world.Position.Get(playerID)
	playerRadius := 0.5

	wallMinX := 50 - 2.0
	wallMaxX := 50 + 2.0
	wallMinY := 50 - 2.0
	wallMaxY := 50 + 2.0

	insideX := pos.X > wallMinX+playerRadius && pos.X < wallMaxX-playerRadius
	insideY := pos.Y > wallMinY+playerRadius && pos.Y < wallMaxY-playerRadius

	if insideX && insideY {
		t.Errorf("Player should be pushed out of wall, still at (%f, %f)", pos.X, pos.Y)
	}
}

func TestCornerCollision_CircleAABB(t *testing.T) {
	world, playerID := setupTestWorld(state.Position{X: 53.5, Y: 53.5}, 0)
	addWall(world, 50, 50, 2, 2)
	ms := NewBasicMovementSystem()

	inputs := map[state.EntityID]PlayerInput{
		playerID: {MoveHorizontal: -1, MoveVertical: -1, MovementType: MovementTypeAbsolute},
	}

	for i := 0; i < 30; i++ {
		ms.Update(1.0/60.0, world, inputs)
		world.ApplyCommands()
	}

	pos, _ := world.Position.Get(playerID)
	playerRadius := 0.5

	wallCornerX := 50 + 2.0
	wallCornerY := 50 + 2.0

	distToCorner := math.Sqrt(
		(pos.X-wallCornerX)*(pos.X-wallCornerX) +
			(pos.Y-wallCornerY)*(pos.Y-wallCornerY),
	)

	if distToCorner < playerRadius-1e-6 {
		t.Errorf("Player too close to corner: dist=%f, radius=%f, pos=(%f,%f)",
			distToCorner, playerRadius, pos.X, pos.Y)
	}
}
