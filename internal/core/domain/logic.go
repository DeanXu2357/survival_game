package domain

import (
	"log"

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

func (gl *Logic) Update(state *State, playerInputs map[string]ports.PlayerInput, dt float64) {
	for playerID, input := range playerInputs {
		// OPTIMIZE: If user react moving sticky we can use Wall Sliding to make it more smooth
		player := state.Players[playerID]

		// Store old values for comparison
		oldDirection := player.Direction
		oldPosition := player.Position

		// Handle rotation
		player.UpdateRotation(&input, dt)

		// Handle movement
		moveVector := player.Move(&input, dt)
		desiredPosition := player.Position.Add(moveVector)

		// Movement processing (debug logs removed)

		for i := 0; i < maxResolutionIteration; i++ {
			collisionOccurred := false

			// Query the spatial grid using the player's circular bounding box to retrieve potentially intersecting objects
			nearObjects := state.ObjectGrid.NearbyPositions(
				vector.Vector2D{max(0, desiredPosition.X-player.Radius), max(0, desiredPosition.Y-player.Radius)},
				vector.Vector2D{max(0, desiredPosition.X-player.Radius), desiredPosition.Y + player.Radius},
				vector.Vector2D{desiredPosition.X + player.Radius, max(0, desiredPosition.Y-player.Radius)},
				vector.Vector2D{desiredPosition.X + player.Radius, desiredPosition.Y + player.Radius},
			)

			// Narrow Phase
			for _, obj := range nearObjects {
				if obj.IsRectangle() {
					isCollisionOccurred, mtv := detectCircleAABBCollision(obj, desiredPosition, player.Radius, moveVector)
					if isCollisionOccurred {
						collisionOccurred = true
						desiredPosition = desiredPosition.Add(mtv)
					}
				}
			}

			if !collisionOccurred {
				break
			}
		}

		// DEBUG: Log the final desired position after collision resolution
		// Position updated
		if oldDirection != player.Direction {
			log.Printf("Player %s rotated from %.2f to %.2f", player.ID, oldDirection, player.Direction)
		}
		if oldPosition != desiredPosition {
			log.Printf("Player %s moved from (%.2f, %.2f) to (%.2f, %.2f)", player.ID, oldPosition.X, oldPosition.Y, desiredPosition.X, desiredPosition.Y)
		}

		player.Position = desiredPosition
	}
}

func detectCircleAABBCollision(obj MapObject, desiredPosition vector.Vector2D, radius float64, moveVector vector.Vector2D) (bool, vector.Vector2D) {
	boundingMin, boundingMax := obj.BoundingBox()

	closestPoint := vector.Vector2D{
		X: max(boundingMin.X, min(desiredPosition.X, boundingMax.X)),
		Y: max(boundingMin.Y, min(desiredPosition.Y, boundingMax.Y)),
	}

	vector2Center := desiredPosition.Sub(closestPoint)

	distanceSquared := vector2Center.X*vector2Center.X + vector2Center.Y*vector2Center.Y

	if distanceSquared < EPSILON {
		// this for temporary fix for zero distance
		// it will return a small movement vector to avoid infinite loop
		// in fact I should implement a wall-sliding algorithm here, but I still study it.
		if moveVector.Magnitude() > EPSILON {
			return true, moveVector.Normalize().Scale(-radius * 0.05)
		}
		return true, vector.Vector2D{0, -1}.Scale(radius * 0.1)
	}

	if distanceSquared < radius*radius { // is colliding
		penetration := radius - desiredPosition.DistanceTo(closestPoint)
		return true, vector2Center.Normalize().Scale(penetration)
	}
	return false, vector.Vector2D{}
}
