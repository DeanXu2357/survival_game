package engine

import (
	"math"
	"testing"

	"survival/internal/engine/ports"
)

// TestGameTickTracking verifies that the tick counter increments correctly
// and that elapsed time is calculated accurately from ticks.
func TestGameTickTracking(t *testing.T) {
	game, err := NewGame(DefaultMapConfig())
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	game.StartGameLoop()

	// Run 60 frames at 1/60 second each (simulating 1 second of game time)
	for i := 0; i < 60; i++ {
		game.Update(1.0 / 60.0)
	}

	// Should have exactly 60 ticks
	if game.CurrentTick() != ports.Tick(60) {
		t.Errorf("Expected 60 ticks, got %d", game.CurrentTick())
	}

	// ElapsedSeconds should be ~1.0 second
	elapsed := game.ElapsedSeconds()
	expected := 1.0
	if math.Abs(elapsed-expected) > 0.01 {
		t.Errorf("Expected ~1.0 second elapsed, got %f", elapsed)
	}
}

// TestTickDeterminism verifies that tick counter is deterministic
// regardless of dt variations (lag spikes, etc.)
func TestTickDeterminism(t *testing.T) {
	game, err := NewGame(DefaultMapConfig())
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	game.StartGameLoop()

	// Pass varying dt values (simulating lag spikes)
	game.Update(1.0 / 60.0) // Normal frame
	game.Update(10.0)       // Huge lag spike (will be clamped to MaxFrameTime)
	game.Update(1.0 / 60.0) // Normal frame

	// Tick counter should be exactly 3, regardless of dt values
	if game.CurrentTick() != ports.Tick(3) {
		t.Errorf("Expected exactly 3 ticks, got %d", game.CurrentTick())
	}

	// Elapsed time should be 3 ticks = 3/60 = 0.05 seconds
	elapsed := game.ElapsedSeconds()
	expected := 3.0 / 60.0
	if math.Abs(elapsed-expected) > 0.001 {
		t.Errorf("Expected %f seconds elapsed, got %f", expected, elapsed)
	}
}

// TestAutoInitialization verifies that Update auto-initializes if StartGameLoop wasn't called
func TestAutoInitialization(t *testing.T) {
	game, err := NewGame(DefaultMapConfig())
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Don't call StartGameLoop explicitly
	if game.IsInitialized() {
		t.Error("Game should not be initialized before first Update()")
	}

	// First Update should auto-initialize
	game.Update(1.0 / 60.0)

	if !game.IsInitialized() {
		t.Error("Game should be auto-initialized after first Update()")
	}

	if game.CurrentTick() != ports.Tick(1) {
		t.Errorf("Expected 1 tick after first Update, got %d", game.CurrentTick())
	}
}

// TestDtClamping verifies that excessive dt values are clamped to MaxFrameTime
func TestDtClamping(t *testing.T) {
	game, err := NewGame(DefaultMapConfig())
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	game.StartGameLoop()

	// Pass an extremely large dt value
	game.Update(100.0) // 100 seconds (would cause physics to explode)

	// Only 1 tick should have passed
	if game.CurrentTick() != ports.Tick(1) {
		t.Errorf("Expected 1 tick, got %d", game.CurrentTick())
	}

	// The dt should have been clamped to MaxFrameTime internally
	// We can't directly test this, but we can verify the game didn't crash
	// and the tick count is correct
}

// TestElapsedSecondsCalculation verifies the conversion from ticks to seconds
func TestElapsedSecondsCalculation(t *testing.T) {
	game, err := NewGame(DefaultMapConfig())
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	game.StartGameLoop()

	testCases := []struct {
		ticks    int
		expected float64
	}{
		{0, 0.0},
		{60, 1.0},
		{120, 2.0},
		{30, 0.5},
		{1, 1.0 / 60.0},
	}

	for _, tc := range testCases {
		// Reset game and run specified number of ticks
		game.StartGameLoop()
		for i := 0; i < tc.ticks; i++ {
			game.Update(1.0 / 60.0)
		}

		elapsed := game.ElapsedSeconds()
		if math.Abs(elapsed-tc.expected) > 0.001 {
			t.Errorf("For %d ticks: expected %f seconds, got %f", tc.ticks, tc.expected, elapsed)
		}
	}
}
