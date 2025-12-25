package domain

import "survival/internal/core/ports"

type MovementSystem struct {
}

func NewMovementSystem() *MovementSystem {
	return &MovementSystem{}
}

// TODO: define how to comunicate with front end about player identity, use string id or entity id ?
func (ms *MovementSystem) Update(dt float64, world *World, buf *CommandBuffer, playerInput map[string]ports.PlayerInput) map[EntityID]Position {
	// TODO: to be implemented
	panic("not implemented")
	return nil
}
