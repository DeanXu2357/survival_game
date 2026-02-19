package state

import "survival/internal/engine/vector"

type Collider struct {
	// Center deprecated
	Center    Position // TODO: refactor to vector2D offset design
	HalfSize  vector.Vector2D
	Direction Direction

	ShapeType ColliderShape
	Radius    float64
	Offset    vector.Vector2D
}

type ColliderShape uint8

const (
	ColliderShapeNone ColliderShape = iota
	ColliderCircle
	ColliderBox
)

func (w Collider) BoundingBox() (min vector.Vector2D, max vector.Vector2D) {
	if w.ShapeType == ColliderBox {
		return vector.Vector2D{
				X: w.Center.X - w.HalfSize.X,
				Y: w.Center.Y - w.HalfSize.Y,
			}, vector.Vector2D{
				X: w.Center.X + w.HalfSize.X,
				Y: w.Center.Y + w.HalfSize.Y,
			}
	}

	// TODO: generate circle bounding box
	//if w.ShapeType == ColliderCircle {
	//}
	return vector.Vector2D{}, vector.Vector2D{}
}
