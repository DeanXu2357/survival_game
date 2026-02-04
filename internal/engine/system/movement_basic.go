package system

import (
	"math"

	"survival/internal/engine/state"
	"survival/internal/engine/vector"
)

type BasicMovementSystem struct {
	world *state.World
}

func NewBasicMovementSystem(world *state.World) *BasicMovementSystem {
	return &BasicMovementSystem{world: world}
}

func (ms *BasicMovementSystem) ReadMeta() state.Meta {
	return state.ComponentInput | state.ComponentPosition | state.ComponentDirection |
		state.ComponentMovementSpeed | state.ComponentRotationSpeed | state.ComponentPlayerHitbox
}

func (ms *BasicMovementSystem) WriteMeta() state.Meta {
	return state.ComponentPosition | state.ComponentDirection | state.ComponentPrePosition | state.ComponentPlayerHitbox
}

func (ms *BasicMovementSystem) Update(dt float64) {
	world := ms.world
	requiredMeta := ms.ReadMeta()

	for entityID, meta := range world.EntityMeta.All() {
		if !meta.Has(requiredMeta) {
			continue
		}

		input, inputExist := world.Input.Get(entityID)
		if !inputExist {
			continue
		}

		moveSpeed, moveSpeedExist := world.MovementSpeed.Get(entityID)
		if !moveSpeedExist {
			continue
		}

		pos, posExist := world.Position.Get(entityID)
		dir, dirExist := world.Direction.Get(entityID)
		rotSpeed, rotSpeedExist := world.RotationSpeed.Get(entityID)
		playerShape, playerShapeExist := world.PlayerHitbox.Get(entityID)

		if !posExist || !dirExist || !rotSpeedExist || !playerShapeExist {
			continue
		}

		var updateMeta state.Meta

		prePos := state.PrePosition(pos)

		newPos := ms.resolvePlayerCollisions(
			ms.calculatePlayerNewPosition(pos, dir, moveSpeed, input, dt),
			playerShape.Radius,
			world,
		)
		newDir := ms.calculatePlayerNewDirection(dir, rotSpeed, input, dt)

		updateMeta = updateMeta.Set(state.ComponentPrePosition)

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
			PrePosition:   prePos,
		})
	}
}

// calculatePlayerNewPosition computes new position based on analog input.
// Uses screen coordinates: Y increases downward.
// MoveVertical: Positive = down, Negative = up
// MoveHorizontal: Positive = right, Negative = left
func (ms *BasicMovementSystem) calculatePlayerNewPosition(pos state.Position, dir state.Direction, speed state.MovementSpeed, input state.Input, dt float64) state.Position {
	var moveX, moveY float64

	switch input.MovementType {
	case state.MovementTypeRelative:
		forward := input.MoveVertical
		strafe := input.MoveHorizontal

		dirRad := float64(dir)
		cosDir := math.Cos(dirRad)
		sinDir := math.Sin(dirRad)

		fwdX := sinDir
		fwdY := -cosDir
		rightX := cosDir
		rightY := sinDir

		moveX = (forward * fwdX) + (strafe * rightX)
		moveY = (forward * fwdY) + (strafe * rightY)

	default:
		moveX = input.MoveHorizontal
		moveY = input.MoveVertical
	}

	movement := vector.Vector2D{X: moveX, Y: moveY}
	if movement.X != 0 || movement.Y != 0 {
		movement = movement.Normalize().Scale(float64(speed) * dt)
	}

	return state.Position(vector.Vector2D(pos).Add(movement))
}

// calculatePlayerNewDirection computes new direction based on rotation input.
// LookHorizontal: Positive = clockwise (right), Negative = counter-clockwise (left).
func (ms *BasicMovementSystem) calculatePlayerNewDirection(dir state.Direction, speed state.RotationSpeed, input state.Input, dt float64) state.Direction {
	rotationDelta := input.LookHorizontal * float64(speed) * dt
	return state.Direction(float64(dir) + rotationDelta)
}

// resolvePlayerCollisions checks for wall collisions and adjusts position.
// Assumes circular player hitbox.
// Uses Circle-AABB collision detection.
func (ms *BasicMovementSystem) resolvePlayerCollisions(pos state.Position, radius float64, world *state.World) state.Position {
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
			wallMin, wallMax := wallShape.BoundingBox() // assume box aligned with axis
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
