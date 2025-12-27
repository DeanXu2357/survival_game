package domain

import (
	"math"

	"survival/internal/core/domain/vector"
	"survival/internal/core/ports"
)

type MovementSystem struct {
}

func NewMovementSystem() *MovementSystem {
	return &MovementSystem{}
}

// Update processes player inputs and updates positions/directions.
// Returns a map of position deltas for downstream systems (vision, etc.).
func (ms *MovementSystem) Update(dt float64, world *World, playerInputs map[EntityID]ports.PlayerInput) map[EntityID]Position {
	positionDeltas := make(map[EntityID]Position)

	for entityID, input := range playerInputs {
		pos, posExist := world.Position.Get(entityID)
		dir, dirExist := world.Direction.Get(entityID)
		moveSpeed, moveSpeedExist := world.MovementSpeed.Get(entityID)
		rotSpeed, rotSpeedExist := world.RotationSpeed.Get(entityID)
		playerShape, playerShapeExist := world.PlayerShape.Get(entityID)

		if !posExist || !dirExist || !moveSpeedExist || !rotSpeedExist || !playerShapeExist {
			// TODO: log error
			continue
		}

		var updateMeta Meta

		newPos := resolvePlayerCollisions(
			calculatePlayerNewPosition(pos, moveSpeed, input, dt),
			playerShape.Radius,
			world,
		)
		newDir := calculatePlayerNewDirection(dir, rotSpeed, input, dt)

		positionDeltas[entityID] = newPos

		if newPos != pos {
			updateMeta = updateMeta.Set(ComponentPosition)
		}
		if newDir != dir {
			updateMeta = updateMeta.Set(ComponentDirection)
		}

		world.UpdatePlayer(entityID, UpdatePlayer{
			UpdateMeta:    updateMeta,
			Position:      newPos,
			Direction:     newDir,
			MovementSpeed: moveSpeed,
			RotationSpeed: rotSpeed,
			PlayerShape:   PlayerShape{Center: newPos, Radius: playerShape.Radius},
		})
	}

	return positionDeltas
}

// calculatePlayerNewPosition computes new position based on WASD input.
// Uses screen coordinates: Y increases downward.
// MoveUp decreases Y, MoveDown increases Y.
func calculatePlayerNewPosition(pos Position, speed MovementSpeed, input ports.PlayerInput, dt float64) Position {
	var moveX, moveY float64
	if input.MoveUp {
		moveY -= 1
	}
	if input.MoveDown {
		moveY += 1
	}
	if input.MoveLeft {
		moveX -= 1
	}
	if input.MoveRight {
		moveX += 1
	}

	movement := vector.Vector2D{X: moveX, Y: moveY}
	if movement.X != 0 || movement.Y != 0 {
		movement = movement.Normalize().Scale(float64(speed) * dt)
	}

	return Position(vector.Vector2D(pos).Add(movement))
}

// calculatePlayerNewDirection computes new direction based on rotation input.
// RotateLeft increases angle (counter-clockwise), RotateRight decreases angle (clockwise).
func calculatePlayerNewDirection(dir Direction, speed RotationSpeed, input ports.PlayerInput, dt float64) Direction {
	var rotationDelta float64
	if input.RotateLeft {
		rotationDelta += float64(speed) * dt
	}
	if input.RotateRight {
		rotationDelta -= float64(speed) * dt
	}
	return Direction(float64(dir) + rotationDelta)
}

// resolvePlayerCollisions checks for wall collisions and adjusts position.
// Uses Circle-AABB collision detection.
func resolvePlayerCollisions(pos Position, radius float64, world *World) Position {
	result := vector.Vector2D(pos)

	playerBounds := Bounds{
		MinX: result.X - radius,
		MinY: result.Y - radius,
		MaxX: result.X + radius,
		MaxY: result.Y + radius,
	}

	for _, cell := range world.Grid.CellsInBounds(playerBounds) {
		for _, entry := range cell.entries {
			if !entry.Layer.Has(LayerStatic) {
				continue
			}

			wallShape, exist := world.WallShape.Get(entry.EntityID)
			if !exist {
				continue
			}
			wallMin, wallMax := wallShape.BoundingBox()
			collides, pushOut := circleAABBCollision(Position(result), radius, wallMin, wallMax)
			if collides {
				result = result.Add(pushOut)
			}
		}
	}

	return Position(result)
}

// circleAABBCollision detects collision between a circle and an AABB.
// Returns whether collision occurred and the push-out vector to resolve it.
func circleAABBCollision(circleCenter Position, radius float64, wallMin, wallMax vector.Vector2D) (collides bool, pushOut vector.Vector2D) {
	center := vector.Vector2D(circleCenter)

	closest := vector.Vector2D{
		X: math.Max(wallMin.X, math.Min(center.X, wallMax.X)),
		Y: math.Max(wallMin.Y, math.Min(center.Y, wallMax.Y)),
	}

	diff := center.Sub(closest)
	dist := diff.Magnitude()

	if dist >= radius {
		return false, vector.Vector2D{}
	}

	if dist > 0 {
		penetration := radius - dist
		return true, diff.Normalize().Scale(penetration)
	}

	// Circle center inside AABB - find nearest edge and push out
	distToLeft := center.X - wallMin.X
	distToRight := wallMax.X - center.X
	distToTop := center.Y - wallMin.Y
	distToBottom := wallMax.Y - center.Y

	minDist := distToLeft
	pushOut = vector.Vector2D{X: -(distToLeft + radius), Y: 0}

	if distToRight < minDist {
		minDist = distToRight
		pushOut = vector.Vector2D{X: distToRight + radius, Y: 0}
	}
	if distToTop < minDist {
		minDist = distToTop
		pushOut = vector.Vector2D{X: 0, Y: -(distToTop + radius)}
	}
	if distToBottom < minDist {
		pushOut = vector.Vector2D{X: 0, Y: distToBottom + radius}
	}

	return true, pushOut
}
