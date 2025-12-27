package domain

import (
	"survival/internal/core/domain/state"
	"survival/internal/core/domain/vector"
)

// PlayerSnapshot represents essential player data for client updates
type PlayerSnapshot struct {
	ID        state.EntityID  `json:"id"`
	Position  vector.Vector2D `json:"position"`
	Direction float64         `json:"direction"`
	Health    int             `json:"health"`
	IsAlive   bool            `json:"isAlive"`
}

// ProjectileSnapshot represents essential projectile data for client updates
type ProjectileSnapshot struct {
	ID        string          `json:"id"`
	Position  vector.Vector2D `json:"position"`
	Direction vector.Vector2D `json:"direction"`
	Speed     float64         `json:"speed"`
}

// GameUpdate is the lightweight message sent to clients during gameplay
type GameUpdate struct {
	Type        string                    `json:"type"`
	Players     map[string]PlayerSnapshot `json:"players"`
	Projectiles []ProjectileSnapshot      `json:"projectiles"`
	Timestamp   int64                     `json:"timestamp"`
}

// StaticGameData contains data sent once on connection (walls, map layout, etc.)
type StaticGameData struct {
	// TODO: define this
}
