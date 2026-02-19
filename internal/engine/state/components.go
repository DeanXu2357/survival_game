package state

import "survival/internal/engine/vector"

type Position vector.Vector2D

type Direction float64

type PrePosition Position

type PlayerHitbox struct {
	Center Position // TODO: refactor to vector2D offset design
	Radius float64
}

type MovementSpeed float64

type RotationSpeed float64

type Health int

type ViewIDs []EntityID

type VerticalBody struct {
	BaseElevation float64
	Height        float64
}
