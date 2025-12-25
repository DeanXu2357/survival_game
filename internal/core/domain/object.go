package domain

import (
	"math"

	"survival/internal/core/domain/vector"
)

/*
 * No sure still needed after refactor, but keeping for now
 */

type Wall struct {
	id       string
	Center   vector.Vector2D `json:"center"`
	HalfSize vector.Vector2D `json:"half_size"`
	Rotation float64         `json:"rotation"`
}

func (w *Wall) GetID() string {
	return w.id
}

func (w *Wall) BoundingBox() (vector.Vector2D, vector.Vector2D) {
	halfX := w.HalfSize.X
	halfY := w.HalfSize.Y

	cos := math.Cos(w.Rotation)
	sin := math.Sin(w.Rotation)

	extentX := math.Abs(halfX*cos) + math.Abs(halfY*sin)
	extentY := math.Abs(halfX*sin) + math.Abs(halfY*cos)

	return vector.Vector2D{
			X: w.Center.X - extentX,
			Y: w.Center.Y - extentY,
		}, vector.Vector2D{
			X: w.Center.X + extentX,
			Y: w.Center.Y + extentY,
		}
}

func (w *Wall) IsRectangle() bool {
	return true
}

func (w *Wall) ID() string {
	return w.id
}

func (w *Wall) Position() vector.Vector2D {
	return w.Center
}

func NewWall(id string, center vector.Vector2D, halfSize vector.Vector2D, rotation float64) *Wall {
	return &Wall{
		id:       id,
		Center:   center,
		HalfSize: halfSize,
		Rotation: rotation,
	}
}
