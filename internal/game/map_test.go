package game

import (
	"testing"
)

// Validation tests moved to individual loader test files

func TestMapConfig_GetRandomSpawnPoint(t *testing.T) {
	config := &MapConfig{
		SpawnPoints: []SpawnPoint{
			{ID: "spawn_1", Position: Vector2D{X: 100, Y: 100}},
			{ID: "spawn_2", Position: Vector2D{X: 200, Y: 200}},
		},
	}

	spawnPoint := config.GetRandomSpawnPoint()
	if spawnPoint == nil {
		t.Error("GetRandomSpawnPoint() returned nil")
		return
	}

	// For now, it should return the first spawn point
	if spawnPoint.ID != "spawn_1" {
		t.Errorf("GetRandomSpawnPoint() returned %s, want spawn_1", spawnPoint.ID)
	}
}

func TestMapConfig_GetRandomSpawnPoint_Empty(t *testing.T) {
	config := &MapConfig{
		SpawnPoints: []SpawnPoint{},
	}

	spawnPoint := config.GetRandomSpawnPoint()
	if spawnPoint != nil {
		t.Error("GetRandomSpawnPoint() should return nil for empty spawn points")
	}
}
