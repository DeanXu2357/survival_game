package raycast

import (
	"math"
	"testing"

	"survival/internal/engine/ports"
)

func TestRayCircleIntersect_HitsAhead(t *testing.T) {
	circle := ports.Collider{X: 5, Y: 0, Radius: 1, ShapeType: shapeCircle}
	dist, hit := rayCircleIntersect(0, 0, 1, 0, circle)
	if !hit {
		t.Fatalf("expected hit")
	}
	if math.Abs(dist-4.0) > 1e-6 {
		t.Fatalf("expected dist ≈ 4.0, got %v", dist)
	}
}

func TestRayCircleIntersect_MissesToSide(t *testing.T) {
	circle := ports.Collider{X: 5, Y: 3, Radius: 1, ShapeType: shapeCircle}
	_, hit := rayCircleIntersect(0, 0, 1, 0, circle)
	if hit {
		t.Fatalf("expected miss (circle offset by 3 on Y, radius 1)")
	}
}

func TestRayCircleIntersect_BehindOrigin(t *testing.T) {
	circle := ports.Collider{X: -5, Y: 0, Radius: 1, ShapeType: shapeCircle}
	_, hit := rayCircleIntersect(0, 0, 1, 0, circle)
	if hit {
		t.Fatalf("expected miss (circle behind ray origin)")
	}
}

func TestCastSingleRay_DispatchesOnShape(t *testing.T) {
	colliders := []ports.Collider{
		{X: 10, Y: 0, HalfX: 1, HalfY: 1, ShapeType: 2},
		{X: 3, Y: 0, Radius: 0.5, ShapeType: shapeCircle},
	}
	dist, hit, hitCol := castSingleRay(0, 0, 1, 0, colliders)
	if !hit {
		t.Fatalf("expected hit")
	}
	if hitCol == nil || hitCol.ShapeType != shapeCircle {
		t.Fatalf("expected circle to be nearest hit, got %+v", hitCol)
	}
	if math.Abs(dist-2.5) > 1e-6 {
		t.Fatalf("expected dist ≈ 2.5, got %v", dist)
	}
}
