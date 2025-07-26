package game

import (
	"math"
	"testing"
)

func TestWall_BoundingBox(t *testing.T) {
	// Helper function to compare vectors with a tolerance
	vectorsAlmostEqual := func(v1, v2 Vector2D, tolerance float64) bool {
		return math.Abs(v1.X-v2.X) < tolerance && math.Abs(v1.Y-v2.Y) < tolerance
	}

	tests := []struct {
		name        string
		wall        *Wall
		expectedMin Vector2D
		expectedMax Vector2D
	}{
		{
			name:        "no rotation",
			wall:        NewWall("wall1", Vector2D{X: 10, Y: 20}, Vector2D{X: 5, Y: 10}, 0),
			expectedMin: Vector2D{X: 5, Y: 10},
			expectedMax: Vector2D{X: 15, Y: 30},
		},
		{
			name:        "90 degree rotation",
			wall:        NewWall("wall2", Vector2D{X: 10, Y: 20}, Vector2D{X: 5, Y: 10}, math.Pi/2),
			expectedMin: Vector2D{X: 0, Y: 15},
			expectedMax: Vector2D{X: 20, Y: 25},
		},
		{
			name:        "45 degree rotation",
			wall:        NewWall("wall3", Vector2D{X: 0, Y: 0}, Vector2D{X: 10, Y: 5}, math.Pi/4),
			expectedMin: Vector2D{X: -(15 / math.Sqrt(2)), Y: -(15 / math.Sqrt(2))},
			expectedMax: Vector2D{X: 15 / math.Sqrt(2), Y: 15 / math.Sqrt(2)},
		},
		{
			name:        "center at origin, no rotation",
			wall:        NewWall("wall4", Vector2D{X: 0, Y: 0}, Vector2D{X: 10, Y: 10}, 0),
			expectedMin: Vector2D{X: -10, Y: -10},
			expectedMax: Vector2D{X: 10, Y: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minVector, maxVector := tt.wall.BoundingBox()
			// Use a small tolerance for float comparisons
			const tolerance = 1e-9

			if !vectorsAlmostEqual(minVector, tt.expectedMin, tolerance) {
				t.Errorf("BoundingBox() minVector = %v, want %v", minVector, tt.expectedMin)
			}
			if !vectorsAlmostEqual(maxVector, tt.expectedMax, tolerance) {
				t.Errorf("BoundingBox() maxVector = %v, want %v", maxVector, tt.expectedMax)
			}
		})
	}
}
