package game

import (
	"math"
	"testing"
)

func TestVector2D_Add(t *testing.T) {
	tests := []struct {
		name     string
		v1       Vector2D
		v2       Vector2D
		expected Vector2D
	}{
		{
			name:     "positive vectors",
			v1:       Vector2D{X: 1.0, Y: 2.0},
			v2:       Vector2D{X: 3.0, Y: 4.0},
			expected: Vector2D{X: 4.0, Y: 6.0},
		},
		{
			name:     "negative vectors",
			v1:       Vector2D{X: -1.0, Y: -2.0},
			v2:       Vector2D{X: -3.0, Y: -4.0},
			expected: Vector2D{X: -4.0, Y: -6.0},
		},
		{
			name:     "mixed vectors",
			v1:       Vector2D{X: 1.5, Y: -2.5},
			v2:       Vector2D{X: -0.5, Y: 3.5},
			expected: Vector2D{X: 1.0, Y: 1.0},
		},
		{
			name:     "zero vector",
			v1:       Vector2D{X: 5.0, Y: 7.0},
			v2:       Vector2D{X: 0.0, Y: 0.0},
			expected: Vector2D{X: 5.0, Y: 7.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Add(tt.v2)
			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("Add() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVector2D_Sub(t *testing.T) {
	tests := []struct {
		name     string
		v1       Vector2D
		v2       Vector2D
		expected Vector2D
	}{
		{
			name:     "positive vectors",
			v1:       Vector2D{X: 5.0, Y: 7.0},
			v2:       Vector2D{X: 2.0, Y: 3.0},
			expected: Vector2D{X: 3.0, Y: 4.0},
		},
		{
			name:     "negative result",
			v1:       Vector2D{X: 1.0, Y: 2.0},
			v2:       Vector2D{X: 3.0, Y: 4.0},
			expected: Vector2D{X: -2.0, Y: -2.0},
		},
		{
			name:     "same vectors",
			v1:       Vector2D{X: 5.0, Y: 5.0},
			v2:       Vector2D{X: 5.0, Y: 5.0},
			expected: Vector2D{X: 0.0, Y: 0.0},
		},
		{
			name:     "subtract zero",
			v1:       Vector2D{X: 3.0, Y: 4.0},
			v2:       Vector2D{X: 0.0, Y: 0.0},
			expected: Vector2D{X: 3.0, Y: 4.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Sub(tt.v2)
			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("Sub() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVector2D_Scale(t *testing.T) {
	tests := []struct {
		name     string
		v        Vector2D
		scalar   float64
		expected Vector2D
	}{
		{
			name:     "positive scalar",
			v:        Vector2D{X: 2.0, Y: 3.0},
			scalar:   2.0,
			expected: Vector2D{X: 4.0, Y: 6.0},
		},
		{
			name:     "negative scalar",
			v:        Vector2D{X: 2.0, Y: -3.0},
			scalar:   -1.0,
			expected: Vector2D{X: -2.0, Y: 3.0},
		},
		{
			name:     "fractional scalar",
			v:        Vector2D{X: 4.0, Y: 6.0},
			scalar:   0.5,
			expected: Vector2D{X: 2.0, Y: 3.0},
		},
		{
			name:     "zero scalar",
			v:        Vector2D{X: 5.0, Y: 7.0},
			scalar:   0.0,
			expected: Vector2D{X: 0.0, Y: 0.0},
		},
		{
			name:     "identity scalar",
			v:        Vector2D{X: 3.0, Y: 4.0},
			scalar:   1.0,
			expected: Vector2D{X: 3.0, Y: 4.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Scale(tt.scalar)
			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("Scale() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVector2D_Dot(t *testing.T) {
	tests := []struct {
		name     string
		v1       Vector2D
		v2       Vector2D
		expected float64
	}{
		{
			name:     "perpendicular vectors",
			v1:       Vector2D{X: 1.0, Y: 0.0},
			v2:       Vector2D{X: 0.0, Y: 1.0},
			expected: 0.0,
		},
		{
			name:     "parallel vectors",
			v1:       Vector2D{X: 2.0, Y: 3.0},
			v2:       Vector2D{X: 4.0, Y: 6.0},
			expected: 26.0,
		},
		{
			name:     "opposite vectors",
			v1:       Vector2D{X: 1.0, Y: 1.0},
			v2:       Vector2D{X: -1.0, Y: -1.0},
			expected: -2.0,
		},
		{
			name:     "unit vectors",
			v1:       Vector2D{X: 1.0, Y: 0.0},
			v2:       Vector2D{X: 1.0, Y: 0.0},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Dot(tt.v2)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Dot() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVector2D_Magnitude(t *testing.T) {
	tests := []struct {
		name     string
		v        Vector2D
		expected float64
	}{
		{
			name:     "unit vector X",
			v:        Vector2D{X: 1.0, Y: 0.0},
			expected: 1.0,
		},
		{
			name:     "unit vector Y",
			v:        Vector2D{X: 0.0, Y: 1.0},
			expected: 1.0,
		},
		{
			name:     "3-4-5 triangle",
			v:        Vector2D{X: 3.0, Y: 4.0},
			expected: 5.0,
		},
		{
			name:     "zero vector",
			v:        Vector2D{X: 0.0, Y: 0.0},
			expected: 0.0,
		},
		{
			name:     "negative components",
			v:        Vector2D{X: -3.0, Y: -4.0},
			expected: 5.0,
		},
		{
			name:     "square root of 2",
			v:        Vector2D{X: 1.0, Y: 1.0},
			expected: math.Sqrt(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Magnitude()
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Magnitude() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVector2D_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		v        Vector2D
		expected Vector2D
	}{
		{
			name:     "unit vector X",
			v:        Vector2D{X: 1.0, Y: 0.0},
			expected: Vector2D{X: 1.0, Y: 0.0},
		},
		{
			name:     "3-4-5 triangle",
			v:        Vector2D{X: 3.0, Y: 4.0},
			expected: Vector2D{X: 0.6, Y: 0.8},
		},
		{
			name:     "negative components",
			v:        Vector2D{X: -2.0, Y: 0.0},
			expected: Vector2D{X: -1.0, Y: 0.0},
		},
		{
			name:     "zero vector",
			v:        Vector2D{X: 0.0, Y: 0.0},
			expected: Vector2D{X: 0.0, Y: 0.0},
		},
		{
			name:     "diagonal vector",
			v:        Vector2D{X: 1.0, Y: 1.0},
			expected: Vector2D{X: 1.0 / math.Sqrt(2), Y: 1.0 / math.Sqrt(2)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Normalize()
			if math.Abs(result.X-tt.expected.X) > 1e-9 || math.Abs(result.Y-tt.expected.Y) > 1e-9 {
				t.Errorf("Normalize() = %v, want %v", result, tt.expected)
			}

			// Check that normalized vector has magnitude 1 (except for zero vector)
			if tt.v.X != 0.0 || tt.v.Y != 0.0 {
				magnitude := result.Magnitude()
				if math.Abs(magnitude-1.0) > 1e-9 {
					t.Errorf("Normalized vector magnitude = %v, want 1.0", magnitude)
				}
			}
		})
	}
}

func TestVector2D_DistanceTo(t *testing.T) {
	tests := []struct {
		name     string
		v1       Vector2D
		v2       Vector2D
		expected float64
	}{
		{
			name:     "same point",
			v1:       Vector2D{X: 3.0, Y: 4.0},
			v2:       Vector2D{X: 3.0, Y: 4.0},
			expected: 0.0,
		},
		{
			name:     "horizontal distance",
			v1:       Vector2D{X: 0.0, Y: 0.0},
			v2:       Vector2D{X: 5.0, Y: 0.0},
			expected: 5.0,
		},
		{
			name:     "vertical distance",
			v1:       Vector2D{X: 0.0, Y: 0.0},
			v2:       Vector2D{X: 0.0, Y: 3.0},
			expected: 3.0,
		},
		{
			name:     "3-4-5 triangle",
			v1:       Vector2D{X: 0.0, Y: 0.0},
			v2:       Vector2D{X: 3.0, Y: 4.0},
			expected: 5.0,
		},
		{
			name:     "negative coordinates",
			v1:       Vector2D{X: -1.0, Y: -1.0},
			v2:       Vector2D{X: 2.0, Y: 3.0},
			expected: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.DistanceTo(tt.v2)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("DistanceTo() = %v, want %v", result, tt.expected)
			}

			// Test symmetry: distance from A to B should equal distance from B to A
			reverse := tt.v2.DistanceTo(tt.v1)
			if math.Abs(result-reverse) > 1e-9 {
				t.Errorf("Distance not symmetric: %v vs %v", result, reverse)
			}
		})
	}
}
