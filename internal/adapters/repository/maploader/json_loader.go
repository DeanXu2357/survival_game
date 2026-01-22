package maploader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"

	"survival/internal/engine"
)

type JSONMapLoader struct {
	mapsDirectory string
	validator     *validator.Validate
}

func NewJSONMapLoader(directory string) *JSONMapLoader {
	return &JSONMapLoader{
		mapsDirectory: directory,
		validator:     validator.New(),
	}
}

func (j *JSONMapLoader) LoadMap(mapID string) (*engine.MapConfig, error) {
	filename := fmt.Sprintf("%s.json", mapID)
	fullPath := filepath.Join(j.mapsDirectory, filename)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read map file %s: %w", fullPath, err)
	}

	var mapData struct {
		Map engine.MapConfig `json:"map"`
	}

	if err := json.Unmarshal(data, &mapData); err != nil {
		return nil, fmt.Errorf("failed to parse map JSON from %s: %w", fullPath, err)
	}

	if err := j.validator.Struct(&mapData.Map); err != nil {
		return nil, fmt.Errorf("invalid map configuration in %s: %w", fullPath, err)
	}

	return &mapData.Map, nil
}

func (j *JSONMapLoader) ListAvailableMaps() ([]string, error) {
	entries, err := os.ReadDir(j.mapsDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read maps directory %s: %w", j.mapsDirectory, err)
	}

	var maps []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			mapID := strings.TrimSuffix(name, ".json")
			maps = append(maps, mapID)
		}
	}

	return maps, nil
}
