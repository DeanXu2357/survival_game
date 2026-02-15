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

// Forward returns the unit forward vector for a heading angle (radians).
// Convention: 0 = north (up), positive = clockwise, Y-down screen coords.
func Forward(rad float64) Vector2D {
	return Vector2D{X: math.Sin(rad), Y: -math.Cos(rad)}
}

// Right returns the unit right vector for a heading angle (radians).
func Right(rad float64) Vector2D {
	return Vector2D{X: math.Cos(rad), Y: math.Sin(rad)}
}
