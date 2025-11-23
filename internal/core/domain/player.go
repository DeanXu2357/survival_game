package domain

import (
	"survival/internal/core/domain/vector"
	"survival/internal/core/domain/weapons"
	"survival/internal/core/ports"
)

const (
	playerBaseMovementSpeed float64 = 5
	playerBaseRotationSpeed float64 = 2
)

type Player struct {
	ID            string
	Position      vector.Vector2D
	Direction     float64
	Radius        float64
	RotationSpeed float64
	MovementSpeed float64
	Health        int
	IsAlive       bool
	Inventory     *weapons.Inventory
	CurrentWeapon weapons.Weapon
}

func (p *Player) Move(input *ports.PlayerInput, dt float64) vector.Vector2D {
	movementVector := vector.Vector2D{X: 0, Y: 0}

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

func (p *Player) UpdateRotation(input *ports.PlayerInput, dt float64) {
	if input.RotateLeft {
		p.Direction -= p.RotationSpeed * dt
	}
	if input.RotateRight {
		p.Direction += p.RotationSpeed * dt
	}
}
