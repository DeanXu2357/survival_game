package vector

import "math"

type Vector2D struct {
	X float64 `json:"x" validate:"required"`
	Y float64 `json:"y" validate:"required"`
}

func (v Vector2D) Add(other Vector2D) Vector2D {
	return Vector2D{
		X: v.X + other.X,
		Y: v.Y + other.Y,
	}
}

func (v Vector2D) Sub(other Vector2D) Vector2D {
	return Vector2D{
		X: v.X - other.X,
		Y: v.Y - other.Y,
	}
}

func (v Vector2D) Scale(scalar float64) Vector2D {
	return Vector2D{
		X: v.X * scalar,
		Y: v.Y * scalar,
	}
}

func (v Vector2D) Dot(other Vector2D) float64 {
	return v.X*other.X + v.Y*other.Y
}

func (v Vector2D) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector2D) Normalize() Vector2D {
	mag := v.Magnitude()
	if mag == 0 {
		return Vector2D{X: 0, Y: 0}
	}
	return Vector2D{
		X: v.X / mag,
		Y: v.Y / mag,
	}
}

func (v Vector2D) DistanceTo(other Vector2D) float64 {
	return v.Sub(other).Magnitude()
}
