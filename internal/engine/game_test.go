package engine_test

import (
	"math"
	"testing"

	"survival/internal/engine"
	"survival/internal/engine/ports"
	"survival/internal/engine/vector"
)

// TestHeadlessLoop verifies game logic runs independently without rendering
func TestHeadlessLoop(t *testing.T) {
	mapConfig := &engine.MapConfig{
		Dimensions: vector.Vector2D{X: 100, Y: 100},
		GridSize:   10,
		Walls:      []engine.WallConfig{},
		SpawnPoints: []engine.SpawnPoint{
			{Position: vector.Vector2D{X: 10, Y: 10}},
		},
	}

	game, err := engine.NewGame(mapConfig)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	playerID, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("Failed to join player: %v", err)
	}

	initialSnapshot, exists := game.PlayerSnapshotWithLocation(playerID)
	if !exists {
		t.Fatal("Player should exist")
	}
	t.Logf("Initial Pos: (%.2f, %.2f)", initialSnapshot.Player.Position.X, initialSnapshot.Player.Position.Y)

	input := ports.PlayerInput{
		MoveHorizontal: 1.0,
		MoveVertical:   0.0,
		Timestamp:      1,
	}

	dt := 1.0 / 60.0
	for i := 0; i < 60; i++ {
		game.SetPlayerInput(playerID, input)
		game.Update(dt)
	}

	finalSnapshot, _ := game.PlayerSnapshotWithLocation(playerID)
	t.Logf("Final Pos: (%.2f, %.2f)", finalSnapshot.Player.Position.X, finalSnapshot.Player.Position.Y)

	if finalSnapshot.Player.Position.X <= initialSnapshot.Player.Position.X {
		t.Errorf("Player did not move! X remained %.2f", finalSnapshot.Player.Position.X)
	}

	expectedX := initialSnapshot.Player.Position.X + 5.0
	if math.Abs(finalSnapshot.Player.Position.X-expectedX) > 0.1 {
		t.Errorf("Movement accuracy failed. Expected X ~%.2f, Got %.2f", expectedX, finalSnapshot.Player.Position.X)
	}
}

// TestWallCollision verifies physics collision system
func TestWallCollision(t *testing.T) {
	mapConfig := &engine.MapConfig{
		Dimensions: vector.Vector2D{X: 100, Y: 100},
		GridSize:   10,
		Walls: []engine.WallConfig{
			{
				Center:   vector.Vector2D{X: 20, Y: 10},
				HalfSize: vector.Vector2D{X: 5, Y: 5}, // Wall spans X: 15~25
			},
		},
		SpawnPoints: []engine.SpawnPoint{
			{Position: vector.Vector2D{X: 10, Y: 10}},
		},
	}

	game, _ := engine.NewGame(mapConfig)
	pid, _ := game.JoinPlayer()

	input := ports.PlayerInput{MoveHorizontal: 1.0}
	dt := 1.0 / 60.0

	for i := 0; i < 180; i++ {
		game.SetPlayerInput(pid, input)
		game.Update(dt)
	}

	snap, _ := game.PlayerSnapshotWithLocation(pid)
	t.Logf("Collision Test Final Pos: X=%.2f", snap.Player.Position.X)

	if snap.Player.Position.X > 16.0 {
		t.Errorf("Collision failed! Player walked through the wall to %.2f", snap.Player.Position.X)
	}
}
