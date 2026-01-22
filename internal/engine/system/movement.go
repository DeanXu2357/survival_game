package system

import (
	"math"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/vector"
)

// todo: remove dependency on ports in system layer, system should have own communication structs

type MovementSystem struct {
}

func NewMovementSystem() *MovementSystem {
	return &MovementSystem{}
}

// Update processes player inputs and updates positions/directions.
// Returns a map of position deltas for downstream systems (vision, etc.).
func (ms *MovementSystem) Update(dt float64, world *state.World, playerInputs map[state.EntityID]ports.PlayerInput) map[state.EntityID]state.Position {
	positionDeltas := make(map[state.EntityID]state.Position)

	for entityID, input := range playerInputs {
		pos, posExist := world.Position.Get(entityID)
		dir, dirExist := world.Direction.Get(entityID)
		moveSpeed, moveSpeedExist := world.MovementSpeed.Get(entityID)
		rotSpeed, rotSpeedExist := world.RotationSpeed.Get(entityID)
		playerShape, playerShapeExist := world.PlayerHitbox.Get(entityID)

		if !posExist || !dirExist || !moveSpeedExist || !rotSpeedExist || !playerShapeExist {
			// TODO: log error
			continue
		}

		var updateMeta state.Meta

		newPos := resolvePlayerCollisions(
			calculatePlayerNewPosition(pos, moveSpeed, input, dt),
			playerShape.Radius,
			world,
		)
		newDir := calculatePlayerNewDirection(dir, rotSpeed, input, dt)

		positionDeltas[entityID] = newPos

		if newPos != pos {
			updateMeta = updateMeta.Set(state.ComponentPosition)
		}
		if newDir != dir {
			updateMeta = updateMeta.Set(state.ComponentDirection)
		}

		world.UpdatePlayer(entityID, state.UpdatePlayer{
			UpdateMeta:    updateMeta,
			Position:      newPos,
			Direction:     newDir,
			MovementSpeed: moveSpeed,
			RotationSpeed: rotSpeed,
			PlayerHitbox:  state.PlayerHitbox{Center: newPos, Radius: playerShape.Radius},
		})
	}

	return positionDeltas
}

// calculatePlayerNewPosition computes new position based on WASD input.
// Uses screen coordinates: Y increases downward.
// MoveUp decreases Y, MoveDown increases Y.
func calculatePlayerNewPosition(pos state.Position, speed state.MovementSpeed, input ports.PlayerInput, dt float64) state.Position {
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

	return state.Position(vector.Vector2D(pos).Add(movement))
}

// calculatePlayerNewDirection computes new direction based on rotation input.
// RotateLeft increases angle (counter-clockwise), RotateRight decreases angle (clockwise).
func calculatePlayerNewDirection(dir state.Direction, speed state.RotationSpeed, input ports.PlayerInput, dt float64) state.Direction {
	var rotationDelta float64
	if input.RotateLeft {
		rotationDelta += float64(speed) * dt
	}
	if input.RotateRight {
		rotationDelta -= float64(speed) * dt
	}
	return state.Direction(float64(dir) + rotationDelta)
}

// resolvePlayerCollisions checks for wall collisions and adjusts position.
// Uses Circle-AABB collision detection.
func resolvePlayerCollisions(pos state.Position, radius float64, world *state.World) state.Position {
	result := vector.Vector2D(pos)

	playerBounds := state.Bounds{
		MinX: result.X - radius,
		MinY: result.Y - radius,
		MaxX: result.X + radius,
		MaxY: result.Y + radius,
	}

	for _, cell := range world.Grid.CellsInBounds(playerBounds) {
		for _, entry := range cell.Entries {
			if !entry.Layer.Has(state.LayerStatic) {
				continue
			}

			wallShape, exist := world.Collider.Get(entry.EntityID)
			if !exist {
				continue
			}
			wallMin, wallMax := wallShape.BoundingBox()
			collides, pushOut := circleAABBCollision(state.Position(result), radius, wallMin, wallMax)
			if collides {
				result = result.Add(pushOut)
			}
		}
	}

	return state.Position(result)
}

// circleAABBCollision detects collision between a circle and an AABB.
// Returns whether collision occurred and the push-out vector to resolve it.
func circleAABBCollision(circleCenter state.Position, radius float64, wallMin, wallMax vector.Vector2D) (collides bool, pushOut vector.Vector2D) {
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
