package game

import (
	"log"

	"survival/internal/protocol"
)

const (
	targetTickRate = 60.0
	deltaTime      = 1.0 / targetTickRate

	maxResolutionIteration = 5
)

type Logic struct {
}

func NewGameLogic() *Logic {
	return &Logic{}
}

func (gl *Logic) Update(state *State, playerInputs map[string]protocol.PlayerInput, dt float64) {
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
				Vector2D{max(0, desiredPosition.X-player.Radius), max(0, desiredPosition.Y-player.Radius)},
				Vector2D{max(0, desiredPosition.X-player.Radius), desiredPosition.Y + player.Radius},
				Vector2D{desiredPosition.X + player.Radius, max(0, desiredPosition.Y-player.Radius)},
				Vector2D{desiredPosition.X + player.Radius, desiredPosition.Y + player.Radius},
			)

			// Narrow Phase
			for _, obj := range nearObjects {
				if obj.IsRectangle() {
					isCollisionOccurred, mtv := detectCircleAABBCollision(obj, desiredPosition, player.Radius)
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

func detectCircleAABBCollision(obj MapObject, desiredPosition Vector2D, radius float64) (bool, Vector2D) {
	boundingMin, boundingMax := obj.BoundingBox()

	closestPoint := Vector2D{
		X: max(boundingMin.X, min(desiredPosition.X, boundingMax.X)),
		Y: max(boundingMin.Y, min(desiredPosition.Y, boundingMax.Y)),
	}

	vector2Center := desiredPosition.Sub(closestPoint)

	distanceSquared := vector2Center.X*vector2Center.X + vector2Center.Y*vector2Center.Y

	if distanceSquared == 0 {
		// TODO: Handle case where the player is exactly at the center of the object
	}

	if distanceSquared < radius*radius { // is colliding
		penetration := radius - desiredPosition.DistanceTo(closestPoint)
		return true, vector2Center.Normalize().Scale(penetration)
	}
	return false, Vector2D{}
}
