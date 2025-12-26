package domain

import (
	"fmt"

	"survival/internal/core/ports"
)

type Game struct {
	world *World
	buf   *CommandBuffer

	movementSys MovementSystem
}

func NewGame(mapConfig *MapConfig) (*Game, error) {
	gridWidth := int(mapConfig.Dimensions.X / mapConfig.GridSize)
	gridHeight := int(mapConfig.Dimensions.Y / mapConfig.GridSize)

	g := &Game{
		world:       NewWorld(mapConfig.GridSize, gridWidth, gridHeight),
		buf:         NewCommandBuffer(),
		movementSys: *NewMovementSystem(),
	}

	if err := g.loadMapEntities(mapConfig); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *Game) loadMapEntities(mapConfig *MapConfig) error {
	for i, wallCfg := range mapConfig.Walls {
		id, ok := g.world.Entity.Alloc()
		if !ok {
			return fmt.Errorf("failed to allocate entity for wall %d", i)
		}
		g.world.WallShape.Add(id, WallShape{
			Center:   Position{X: wallCfg.Center.X, Y: wallCfg.Center.Y},
			HalfSize: wallCfg.HalfSize,
		})
	}
	return nil
}

func (g *Game) JoinPlayer() (EntityID, error) {
	// TODO: to be implemented - allocate player entity and return its ID
	panic("not implemented")
}

func (g *Game) UpdateInLoop(dt float64, playerInputs map[EntityID]ports.PlayerInput) {
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
