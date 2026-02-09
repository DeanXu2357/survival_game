package session_test

import (
	"testing"

	"survival/internal/engine"
	"survival/internal/engine/ports"
	"survival/internal/engine/vector"
	"survival/internal/terminal/session"
)

func TestGameSession_PlayerStateAtSpawn(t *testing.T) {
	mapConfig := engine.DefaultMapConfig()
	game, err := engine.NewGame(mapConfig)
	if err != nil {
		t.Fatalf("failed to create game: %v", err)
	}

	entityID, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("failed to join player: %v", err)
	}

	sess := session.NewGameSession(game, entityID)

	x, y, dir, ok := sess.PlayerState()
	if !ok {
		t.Fatal("PlayerState returned not ok")
	}

	spawnPoint := mapConfig.SpawnPoints[0]
	if x != spawnPoint.Position.X {
		t.Errorf("expected X=%.1f, got X=%.1f", spawnPoint.Position.X, x)
	}
	if y != spawnPoint.Position.Y {
		t.Errorf("expected Y=%.1f, got Y=%.1f", spawnPoint.Position.Y, y)
	}
	if dir != 0 {
		t.Errorf("expected Dir=0, got Dir=%.2f", dir)
	}
}

func TestGameSession_Movement(t *testing.T) {
	mapConfig := engine.DefaultMapConfig()
	game, err := engine.NewGame(mapConfig)
	if err != nil {
		t.Fatalf("failed to create game: %v", err)
	}

	entityID, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("failed to join player: %v", err)
	}

	sess := session.NewGameSession(game, entityID)

	initialX, initialY, _, ok := sess.PlayerState()
	if !ok {
		t.Fatal("initial PlayerState returned not ok")
	}

	input := ports.PlayerInput{
		MoveVertical: 1,
		MovementType: ports.MovementTypeRelative,
	}
	sess.SetInput(input)

	for i := 0; i < 10; i++ {
		sess.Update(1.0 / 60.0)
	}

	newX, newY, _, ok := sess.PlayerState()
	if !ok {
		t.Fatal("PlayerState after movement returned not ok")
	}

	if newX == initialX && newY == initialY {
		t.Error("player position did not change after movement input")
	}
}

func TestGameSession_Statics(t *testing.T) {
	mapConfig := engine.DefaultMapConfig()
	mapConfig.Walls = append(mapConfig.Walls, engine.WallConfig{
		ID:       "test-wall",
		Center:   vector.Vector2D{X: 100, Y: 100},
		HalfSize: vector.Vector2D{X: 10, Y: 10},
	})

	game, err := engine.NewGame(mapConfig)
	if err != nil {
		t.Fatalf("failed to create game: %v", err)
	}

	entityID, err := game.JoinPlayer()
	if err != nil {
		t.Fatalf("failed to join player: %v", err)
	}

	sess := session.NewGameSession(game, entityID)

	statics := sess.Statics()
	if len(statics) == 0 {
		t.Error("expected at least one static entity")
	}
}
