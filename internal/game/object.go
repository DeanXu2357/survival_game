package game

import "math"

type Wall struct {
	id       string
	Center   Vector2D
	HalfSize Vector2D
	Rotation float64
}

func (w *Wall) BoundingBox() (Vector2D, Vector2D) {
	halfX := w.HalfSize.X
	halfY := w.HalfSize.Y

	cos := math.Cos(w.Rotation)
	sin := math.Sin(w.Rotation)

	extentX := math.Abs(halfX*cos) + math.Abs(halfY*sin)
	extentY := math.Abs(halfX*sin) + math.Abs(halfY*cos)

	return Vector2D{
			X: w.Center.X - extentX,
			Y: w.Center.Y - extentY,
		}, Vector2D{
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

func (w *Wall) Position() Vector2D {
	return w.Center
}

func NewWall(id string, center Vector2D, halfSize Vector2D, rotation float64) *Wall {
	return &Wall{
		id:       id,
		Center:   center,
		HalfSize: halfSize,
		Rotation: rotation,
	}
}
