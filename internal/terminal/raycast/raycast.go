package raycast

import (
	"math"
	"sort"

	"survival/internal/engine/ports"
)

const (
	FOVAngle    = math.Pi / 2
	MaxDistance = 20.0
)

type RaycastResult struct {
	Distance      float64
	Hit           bool
	WallHeight    float64
	BaseElevation float64
	EntityID      uint64
	ShapeType     uint8
}

func CastRays(playerX, playerY, playerDir, viewHeight float64, colliders []ports.Collider, numRays int) [][]RaycastResult {
	results := make([][]RaycastResult, numRays)

	halfFOV := FOVAngle / 2
	startAngle := playerDir - halfFOV
	angleStep := FOVAngle / float64(numRays-1)

	for i := 0; i < numRays; i++ {
		rayAngle := startAngle + float64(i)*angleStep
		rayDirX := math.Sin(rayAngle)
		rayDirY := -math.Cos(rayAngle)

		hits := castSingleRay(playerX, playerY, rayDirX, rayDirY, colliders)

		// Fisheye correction: project each distance onto the view axis.
		angleDiff := rayAngle - playerDir
		cosDiff := math.Cos(angleDiff)
		for j := range hits {
			hits[j].Distance *= cosDiff
		}

		results[i] = hits
	}

	return results
}

// Wire-format values mirror state.ColliderShape (state/collider.go).
const (
	shapeCircle uint8 = 1
)

func castSingleRay(originX, originY, dirX, dirY float64, colliders []ports.Collider) []RaycastResult {
	var hits []RaycastResult

	for i := range colliders {
		var dist float64
		var hit bool
		if colliders[i].ShapeType == shapeCircle {
			dist, hit = rayCircleIntersect(originX, originY, dirX, dirY, colliders[i])
		} else {
			dist, hit = rayBoxIntersect(originX, originY, dirX, dirY, colliders[i])
		}
		if !hit || dist <= 0.001 || dist >= MaxDistance {
			continue
		}
		hits = append(hits, RaycastResult{
			Distance:      dist,
			Hit:           true,
			WallHeight:    colliders[i].Height,
			BaseElevation: colliders[i].BaseElevation,
			EntityID:      colliders[i].ID,
			ShapeType:     colliders[i].ShapeType,
		})
	}

	// Farthest first so the renderer paints back-to-front.
	sort.Slice(hits, func(i, j int) bool {
		return hits[i].Distance > hits[j].Distance
	})

	return hits
}

func rayCircleIntersect(originX, originY, dirX, dirY float64, circle ports.Collider) (float64, bool) {
	ox := originX - circle.X
	oy := originY - circle.Y

	b := ox*dirX + oy*dirY
	c := ox*ox + oy*oy - circle.Radius*circle.Radius

	discriminant := b*b - c
	if discriminant < 0 {
		return 0, false
	}

	sqrtDisc := math.Sqrt(discriminant)
	t := -b - sqrtDisc
	if t < 0 {
		return 0, false
	}

	if t > MaxDistance {
		return MaxDistance, false
	}

	return t, true
}

func rayBoxIntersect(originX, originY, dirX, dirY float64, box ports.Collider) (float64, bool) {
	minX := box.X - box.HalfX
	maxX := box.X + box.HalfX
	minY := box.Y - box.HalfY
	maxY := box.Y + box.HalfY

	var tMin, tMax float64 = -math.MaxFloat64, math.MaxFloat64

	if math.Abs(dirX) < 1e-10 {
		if originX < minX || originX > maxX {
			return 0, false
		}
	} else {
		invDirX := 1.0 / dirX
		t1 := (minX - originX) * invDirX
		t2 := (maxX - originX) * invDirX

		if t1 > t2 {
			t1, t2 = t2, t1
		}

		tMin = math.Max(tMin, t1)
		tMax = math.Min(tMax, t2)
	}

	if math.Abs(dirY) < 1e-10 {
		if originY < minY || originY > maxY {
			return 0, false
		}
	} else {
		invDirY := 1.0 / dirY
		t1 := (minY - originY) * invDirY
		t2 := (maxY - originY) * invDirY

		if t1 > t2 {
			t1, t2 = t2, t1
		}

		tMin = math.Max(tMin, t1)
		tMax = math.Min(tMax, t2)
	}

	if tMax < 0 || tMin > tMax {
		return 0, false
	}

	t := tMin
	if t < 0 {
		t = tMax
	}

	if t > MaxDistance {
		return MaxDistance, false
	}

	return t, true
}
