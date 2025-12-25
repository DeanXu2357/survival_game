package domain

import "survival/internal/core/domain/vector"

type Meta int64

// todo: implement Meta bit mask functionalities

type Position vector.Vector2D

type Direction float64

type PlayerShape struct {
	Center *Position
	Radius float64
}

type MovementSpeed float64

type RotationSpeed float64

type Health int

type WallShape struct {
	Center *Position
	// TODO: define how to get bounding box from wall
}
