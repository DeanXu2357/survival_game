package domain

import "survival/internal/core/ports"

type Game struct {
	world *World
	buf   *CommandBuffer

	movementSys MovementSystem
}

// TODO: to implement initialization from map config
func NewGame() *Game {
	return &Game{}
}

func (g *Game) JoinPlayer() error {
	// todo: to be implemented
	panic("not implemented")
}

func (g *Game) UpdateInLoop(dt float64, playerInputs map[string]ports.PlayerInput) {
	//positionDelta
	_ = g.movementSys.Update(dt, g.world, g.buf, playerInputs)

	// visionDelta := g.visionSys.Update(dt, g.world, g.buf, positionDelta)
	// aliveDelta := g.combatSys.Update(dt, g.world, g.buf)

	g.applyCommands()

	// note: maybe can log here for debug ?
}

func (g *Game) applyCommands() {
	// TODO: to be implemented updating world state from command buffer
}

func (g *Game) Statics() *StaticGameData {
	// TODO: to be implemented
	panic("not implemented")
}
