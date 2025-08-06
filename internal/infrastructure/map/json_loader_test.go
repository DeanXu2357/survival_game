package maploader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJSONMapLoader_LoadMap(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "map_loader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test map file
	testMapJSON := `{
		"map": {
			"id": "test_map",
			"name": "Test Map",
			"dimensions": {"x": 800, "y": 600},
			"grid_size": 50,
			"spawn_points": [
				{
					"id": "spawn_1",
					"position": {"x": 100, "y": 100}
				}
			],
			"walls": [
				{
					"id": "wall_1",
					"center": {"x": 400, "y": 50},
					"half_size": {"x": 400, "y": 25},
					"rotation": 0
				}
			],
			"objects": []
		}
	}`

	mapFilePath := filepath.Join(tempDir, "test_map.json")
	err = os.WriteFile(mapFilePath, []byte(testMapJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write test map file: %v", err)
	}

	loader := NewJSONMapLoader(tempDir)

	// Test successful loading
	mapConfig, err := loader.LoadMap("test_map")
	if err != nil {
		t.Fatalf("LoadMap() failed: %v", err)
	}

	if mapConfig.ID != "test_map" {
		t.Errorf("MapConfig.ID = %s, want test_map", mapConfig.ID)
	}
	if mapConfig.Name != "Test Map" {
		t.Errorf("MapConfig.Name = %s, want Test Map", mapConfig.Name)
	}
	if mapConfig.Dimensions.X != 800 || mapConfig.Dimensions.Y != 600 {
		t.Errorf("MapConfig.Dimensions = %+v, want {X:800, Y:600}", mapConfig.Dimensions)
	}
	if len(mapConfig.SpawnPoints) != 1 {
		t.Errorf("MapConfig.SpawnPoints length = %d, want 1", len(mapConfig.SpawnPoints))
	}
	if len(mapConfig.Walls) != 1 {
		t.Errorf("MapConfig.Walls length = %d, want 1", len(mapConfig.Walls))
	}
}

func TestJSONMapLoader_LoadMap_FileNotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "map_loader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	loader := NewJSONMapLoader(tempDir)

	_, err = loader.LoadMap("nonexistent_map")
	if err == nil {
		t.Error("LoadMap() should fail for nonexistent file")
	}
}

func TestJSONMapLoader_LoadMap_InvalidJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "map_loader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	invalidJSON := `{invalid json`
	mapFilePath := filepath.Join(tempDir, "invalid_map.json")
	err = os.WriteFile(mapFilePath, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid map file: %v", err)
	}

	loader := NewJSONMapLoader(tempDir)

	_, err = loader.LoadMap("invalid_map")
	if err == nil {
		t.Error("LoadMap() should fail for invalid JSON")
	}
}

func TestJSONMapLoader_LoadMap_InvalidMapConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "map_loader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Map with missing required fields
	invalidMapJSON := `{
		"map": {
			"id": "",
			"name": "Invalid Map",
			"dimensions": {"x": 800, "y": 600},
			"grid_size": 50,
			"spawn_points": []
		}
	}`

	mapFilePath := filepath.Join(tempDir, "invalid_config_map.json")
	err = os.WriteFile(mapFilePath, []byte(invalidMapJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid config map file: %v", err)
	}

	loader := NewJSONMapLoader(tempDir)

	_, err = loader.LoadMap("invalid_config_map")
	if err == nil {
		t.Error("LoadMap() should fail for invalid map configuration")
	}
}

func TestJSONMapLoader_ListAvailableMaps(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "map_loader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test map files
	mapFiles := []string{"map1.json", "map2.json", "not_a_map.txt"}
	for _, filename := range mapFiles {
		filePath := filepath.Join(tempDir, filename)
		err = os.WriteFile(filePath, []byte("{}"), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	loader := NewJSONMapLoader(tempDir)

	maps, err := loader.ListAvailableMaps()
	if err != nil {
		t.Fatalf("ListAvailableMaps() failed: %v", err)
	}

	expectedMaps := []string{"map1", "map2"}
	if len(maps) != len(expectedMaps) {
		t.Errorf("ListAvailableMaps() returned %d maps, want %d", len(maps), len(expectedMaps))
	}

	mapSet := make(map[string]bool)
	for _, mapID := range maps {
		mapSet[mapID] = true
	}

	for _, expectedMap := range expectedMaps {
		if !mapSet[expectedMap] {
			t.Errorf("ListAvailableMaps() missing map %s", expectedMap)
		}
	}
}

func TestJSONMapLoader_ListAvailableMaps_DirectoryNotFound(t *testing.T) {
	loader := NewJSONMapLoader("/nonexistent/directory")

	_, err := loader.ListAvailableMaps()
	if err == nil {
		t.Error("ListAvailableMaps() should fail for nonexistent directory")
	}
}
