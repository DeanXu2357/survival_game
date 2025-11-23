package domain

import (
	"survival/internal/core/domain/vector"
)

type MapLoader interface {
	LoadMap(mapID string) (*MapConfig, error)
	ListAvailableMaps() ([]string, error)
}

type MapConfig struct {
	ID          string          `json:"id" validate:"required,min=1"`
	Name        string          `json:"name" validate:"required,min=1"`
	Dimensions  vector.Vector2D `json:"dimensions" validate:"required"`
	GridSize    float64         `json:"grid_size" validate:"required,gt=0"`
	SpawnPoints []SpawnPoint    `json:"spawn_points" validate:"required,min=1,dive"`
	Walls       []WallConfig    `json:"walls" validate:"dive"`
	Objects     []ObjectConfig  `json:"objects,omitempty" validate:"dive"`
}

type SpawnPoint struct {
	ID       string          `json:"id" validate:"required,min=1"`
	Position vector.Vector2D `json:"position" validate:"required"`
}

type WallConfig struct {
	ID       string          `json:"id" validate:"required,min=1"`
	Center   vector.Vector2D `json:"center" validate:"required"`
	HalfSize vector.Vector2D `json:"half_size" validate:"required"`
	Rotation float64         `json:"rotation"`
}

type ObjectConfig struct {
	ID       string          `json:"id" validate:"required,min=1"`
	Type     string          `json:"type" validate:"required,min=1"`
	Center   vector.Vector2D `json:"center" validate:"required"`
	HalfSize vector.Vector2D `json:"half_size" validate:"required"`
	Rotation float64         `json:"rotation"`
}

func (mc *MapConfig) GetRandomSpawnPoint() *SpawnPoint {
	if len(mc.SpawnPoints) == 0 {
		return nil
	}
	// For now, return the first spawn point
	// TODO: implement proper random selection
	return &mc.SpawnPoints[0]
}
