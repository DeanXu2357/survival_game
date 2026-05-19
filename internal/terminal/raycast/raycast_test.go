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

func TestCastSingleRay_ReturnsAllHitsFarthestFirst(t *testing.T) {
	colliders := []ports.Collider{
		{X: 10, Y: 0, HalfX: 1, HalfY: 1, ShapeType: 2},
		{X: 3, Y: 0, Radius: 0.5, ShapeType: shapeCircle},
	}
	hits := castSingleRay(0, 0, 1, 0, colliders)
	if len(hits) != 2 {
		t.Fatalf("expected 2 hits, got %d", len(hits))
	}
	if hits[0].ShapeType != 2 {
		t.Fatalf("expected farthest hit (box) first, got shape %d", hits[0].ShapeType)
	}
	if hits[1].ShapeType != shapeCircle {
		t.Fatalf("expected nearest hit (circle) second, got shape %d", hits[1].ShapeType)
	}
	if math.Abs(hits[1].Distance-2.5) > 1e-6 {
		t.Fatalf("expected circle dist ≈ 2.5, got %v", hits[1].Distance)
	}
	if math.Abs(hits[0].Distance-9.0) > 1e-6 {
		t.Fatalf("expected box dist ≈ 9.0, got %v", hits[0].Distance)
	}
}
