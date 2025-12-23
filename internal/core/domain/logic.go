package domain

import (
	"log"
	"math"

	"survival/internal/core/domain/vector"
	"survival/internal/core/ports"
)

const (
	targetTickRate = 60.0
	deltaTime      = 1.0 / targetTickRate

	maxResolutionIteration = 5
	EPSILON                = 1e-9
)

type Logic struct {
}

func NewGameLogic() *Logic {
	return &Logic{}
}

// Update processes all player inputs and updates the world state.
func (gl *Logic) Update(world *World, playerInputs map[string]ports.PlayerInput, dt float64) {
	for playerID, input := range playerInputs {
		entityID, ok := world.GetEntityByPlayerID(playerID)
		if !ok {
			continue
		}

		gl.handlePlayerMovement(world, entityID, input, dt)

		// TODO: handle interact with objects
	}

	// TODO: handle shooting

	// TODO: handle vision

	// TODO: inform player about state changes
}

func (gl *Logic) handlePlayerMovement(world *World, entityID EntityID, input ports.PlayerInput, dt float64) {
	// Get required components
	pos := world.Positions().Get(entityID)
	dir := world.Directions().Get(entityID)
	stats := world.PlayerStatsManager().Get(entityID)
	collider := world.CircleColliders().Get(entityID)

	if pos == nil || dir == nil || stats == nil || collider == nil {
		return
	}

	// Store old values for comparison
	oldDirection := dir.Angle
	oldPosition := Position{X: pos.X, Y: pos.Y}

	// Handle rotation
	if input.RotateLeft {
		dir.Angle -= stats.RotationSpeed * dt
	}
	if input.RotateRight {
		dir.Angle += stats.RotationSpeed * dt
	}

	// Calculate movement vector
	moveVector := calculateMoveVector(input, stats.MovementSpeed, dt)
	desiredX := pos.X + moveVector.X
	desiredY := pos.Y + moveVector.Y

	// Collision resolution loop
	for i := 0; i < maxResolutionIteration; i++ {
		collisionOccurred := false

		// Query nearby static entities using spatial grid
		searchBounds := Bounds{
			MinX: desiredX - collider.Radius,
			MinY: desiredY - collider.Radius,
			MaxX: desiredX + collider.Radius,
			MaxY: desiredY + collider.Radius,
		}

		for _, cell := range world.Grid().CellsInBounds(searchBounds) {
			for _, entry := range cell.entries {
				if !entry.Layer.Has(LayerStatic) {
					continue
				}

				// Get box collider for static entity
				box := world.BoxColliders().Get(entry.EntityID)
				staticPos := world.Positions().Get(entry.EntityID)
				if box == nil || staticPos == nil {
					continue
				}

				// Check circle-AABB collision
				isCollision, mtv := detectCircleBoxCollision(
					Position{X: desiredX, Y: desiredY},
					collider.Radius,
					*staticPos,
					*box,
					moveVector,
				)

				if isCollision {
					collisionOccurred = true
					desiredX += mtv.X
					desiredY += mtv.Y
				}
			}
		}

		if !collisionOccurred {
			break
		}
	}

	// Update position if changed
	if oldPosition.X != desiredX || oldPosition.Y != desiredY {
		log.Printf("Entity %d moved from (%.2f, %.2f) to (%.2f, %.2f)",
			entityID, oldPosition.X, oldPosition.Y, desiredX, desiredY)

		// Update grid position
		world.UpdateEntityGridPosition(entityID, Position{X: desiredX, Y: desiredY}, collider.Radius, LayerPlayer)

		pos.X = desiredX
		pos.Y = desiredY
	}

	if oldDirection != dir.Angle {
		log.Printf("Entity %d rotated from %.2f to %.2f", entityID, oldDirection, dir.Angle)
	}
}

func calculateMoveVector(input ports.PlayerInput, speed, dt float64) vector.Vector2D {
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

	v := vector.Vector2D{X: moveX, Y: moveY}
	return v.Normalize().Scale(speed * dt)
}

func detectCircleBoxCollision(circlePos Position, radius float64, boxPos Position, box BoxCollider, moveVector vector.Vector2D) (bool, vector.Vector2D) {
	// Calculate AABB bounds for potentially rotated box
	cos := math.Cos(box.Rotation)
	sin := math.Sin(box.Rotation)

	extentX := math.Abs(box.HalfWidth*cos) + math.Abs(box.HalfHeight*sin)
	extentY := math.Abs(box.HalfWidth*sin) + math.Abs(box.HalfHeight*cos)

	boundingMin := vector.Vector2D{X: boxPos.X - extentX, Y: boxPos.Y - extentY}
	boundingMax := vector.Vector2D{X: boxPos.X + extentX, Y: boxPos.Y + extentY}

	// Find closest point on AABB to circle center
	closestX := math.Max(boundingMin.X, math.Min(circlePos.X, boundingMax.X))
	closestY := math.Max(boundingMin.Y, math.Min(circlePos.Y, boundingMax.Y))

	// Vector from closest point to circle center
	vecX := circlePos.X - closestX
	vecY := circlePos.Y - closestY

	distSquared := vecX*vecX + vecY*vecY

	if distSquared < EPSILON {
		// Circle center is inside or very close to AABB
		if moveVector.Magnitude() > EPSILON {
			mv := moveVector.Normalize().Scale(-radius * 0.05)
			return true, mv
		}
		return true, vector.Vector2D{X: 0, Y: -1}.Scale(radius * 0.1)
	}

	if distSquared < radius*radius {
		// Collision detected
		dist := math.Sqrt(distSquared)
		penetration := radius - dist
		normX := vecX / dist
		normY := vecY / dist
		return true, vector.Vector2D{X: normX * penetration, Y: normY * penetration}
	}

	return false, vector.Vector2D{}
}
