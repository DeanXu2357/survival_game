package domain

import "survival/internal/core/domain/vector"

// Position represents the world position of an entity.
type Position struct {
	X float64
	Y float64
}

// ToVector2D converts Position to Vector2D for compatibility.
func (p Position) ToVector2D() vector.Vector2D {
	return vector.Vector2D{X: p.X, Y: p.Y}
}

// PositionFromVector2D creates a Position from Vector2D.
func PositionFromVector2D(v vector.Vector2D) Position {
	return Position{X: v.X, Y: v.Y}
}

// Velocity represents the current velocity of an entity.
type Velocity struct {
	X float64
	Y float64
}

// Direction represents the facing direction (in radians) of an entity.
type Direction struct {
	Angle float64
}

// CircleCollider represents a circular collision shape for players and circular objects.
type CircleCollider struct {
	Radius float64
}

// BoxCollider represents a rectangular collision shape for walls and rectangular objects.
type BoxCollider struct {
	HalfWidth  float64
	HalfHeight float64
	Rotation   float64 // radians
}

// PlayerTag marks an entity as a player (zero-size marker component).
type PlayerTag struct{}

// StaticTag marks an entity as static (walls, obstacles - zero-size marker component).
type StaticTag struct{}

// ProjectileTag marks an entity as a projectile (zero-size marker component).
type ProjectileTag struct{}

// PlayerStats holds player gameplay attributes.
type PlayerStats struct {
	Health        int
	IsAlive       bool
	MovementSpeed float64
	RotationSpeed float64
}

// PlayerIdentity links an entity to network session.
type PlayerIdentity struct {
	SessionID string // WebSocket session ID
}

// GridCells stores which grid cells an entity occupies.
// Used to efficiently remove entity from grid on position update.
type GridCells struct {
	Indexes []int
}
