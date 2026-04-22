package raycast

import (
	"math"

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
}

func CastRays(playerX, playerY, playerDir, viewHeight float64, colliders []ports.Collider, numRays int) []RaycastResult {
	results := make([]RaycastResult, numRays)

	halfFOV := FOVAngle / 2
	startAngle := playerDir - halfFOV
	angleStep := FOVAngle / float64(numRays-1)

	for i := 0; i < numRays; i++ {
		rayAngle := startAngle + float64(i)*angleStep
		rayDirX := math.Sin(rayAngle)
		rayDirY := -math.Cos(rayAngle)

		distance, hit, hitCollider := castSingleRay(playerX, playerY, rayDirX, rayDirY, colliders)

		angleDiff := rayAngle - playerDir
		distance *= math.Cos(angleDiff)

		result := RaycastResult{
			Distance: distance,
			Hit:      hit,
		}

		if hit && hitCollider != nil {
			result.WallHeight = hitCollider.Height
			result.BaseElevation = hitCollider.BaseElevation
			result.EntityID = hitCollider.ID
		}

		results[i] = result
	}

	return results
}

// Wire-format values mirror state.ColliderShape (state/collider.go).
const (
	shapeCircle uint8 = 1
)

func castSingleRay(originX, originY, dirX, dirY float64, colliders []ports.Collider) (float64, bool, *ports.Collider) {
	closestDist := MaxDistance
	hitAnything := false
	var hitCollider *ports.Collider

	for i := range colliders {
		var dist float64
		var hit bool
		if colliders[i].ShapeType == shapeCircle {
			dist, hit = rayCircleIntersect(originX, originY, dirX, dirY, colliders[i])
		} else {
			dist, hit = rayBoxIntersect(originX, originY, dirX, dirY, colliders[i])
		}
		if hit && dist < closestDist && dist > 0.001 {
			closestDist = dist
			hitAnything = true
			hitCollider = &colliders[i]
		}
	}

	return closestDist, hitAnything, hitCollider
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
