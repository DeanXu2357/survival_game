package game

import "survival/internal/protocol"

const (
	playerBaseMovementSpeed float64 = 1
	playerBaseRotationSpeed float64 = 2
)

type Player struct {
	ID            string
	Position      Vector2D
	Direction     float64
	Radius        float64
	RotationSpeed float64
	MovementSpeed float64
	Health        int
	IsAlive       bool
	Inventory     *Inventory
	CurrentWeapon *Weapon
}

func (p *Player) Move(input *protocol.PlayerInput, dt float64) Vector2D {
	movementVector := Vector2D{X: 0, Y: 0}

	if input.MoveUp {
		movementVector.Y += 1
	}
	if input.MoveDown {
		movementVector.Y -= 1
	}
	if input.MoveLeft {
		movementVector.X -= 1
	}
	if input.MoveRight {
		movementVector.X += 1
	}

	return movementVector.Normalize().Scale(p.MovementSpeed * dt)
}
