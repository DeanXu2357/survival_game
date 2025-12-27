package state

import "survival/internal/core/domain/vector"

type Meta uint64

const (
	ComponentPosition Meta = 1 << iota
	ComponentDirection
	ComponentMovementSpeed
	ComponentRotationSpeed
	ComponentPlayerShape
	ComponentHealth
	ComponentWallShape

	PlayerMeta = ComponentPosition | ComponentDirection | ComponentMovementSpeed | ComponentRotationSpeed | ComponentPlayerShape | ComponentHealth
)

func (m Meta) Has(mask Meta) bool {
	return m&mask == mask
}

func (m Meta) Set(mask Meta) Meta {
	return m | mask
}

func (m Meta) Clear(mask Meta) Meta {
	return m &^ mask
}

type Position vector.Vector2D

type Direction float64

type PlayerShape struct {
	Center Position
	Radius float64
}

type MovementSpeed float64

type RotationSpeed float64

type Health int

type WallShape struct {
	Center   Position
	HalfSize vector.Vector2D
}

func (w WallShape) BoundingBox() (min vector.Vector2D, max vector.Vector2D) {
	return vector.Vector2D{
			X: w.Center.X - w.HalfSize.X,
			Y: w.Center.Y - w.HalfSize.Y,
		}, vector.Vector2D{
			X: w.Center.X + w.HalfSize.X,
			Y: w.Center.Y + w.HalfSize.Y,
		}
}
