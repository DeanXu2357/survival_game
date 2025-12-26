package domain

import (
	"fmt"

	"survival/internal/core/ports"
)

type Game struct {
	world     *World
	buf       *CommandBuffer
	mapConfig *MapConfig

	movementSys MovementSystem
}

func NewGame(mapConfig *MapConfig) (*Game, error) {
	gridWidth := int(mapConfig.Dimensions.X / mapConfig.GridSize)
	gridHeight := int(mapConfig.Dimensions.Y / mapConfig.GridSize)

	g := &Game{
		world:       NewWorld(mapConfig.GridSize, gridWidth, gridHeight),
		buf:         NewCommandBuffer(),
		mapConfig:   mapConfig,
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

const (
	defaultPlayerMovementSpeed float64 = 5
	defaultPlayerRotationSpeed float64 = 2
	defaultPlayerRadius        float64 = 0.5
	defaultPlayerHealth        int     = 100
)

func (g *Game) JoinPlayer() (EntityID, error) {
	spawnPoint := g.mapConfig.GetRandomSpawnPoint()
	if spawnPoint == nil {
		return 0, fmt.Errorf("no spawn point available")
	}

	id, ok := g.world.CreatePlayer(PlayerConfig{
		Position:      Position{X: spawnPoint.Position.X, Y: spawnPoint.Position.Y},
		Direction:     0,
		MovementSpeed: MovementSpeed(defaultPlayerMovementSpeed),
		RotationSpeed: RotationSpeed(defaultPlayerRotationSpeed),
		Radius:        defaultPlayerRadius,
		Health:        Health(defaultPlayerHealth),
	})
	if !ok {
		return 0, fmt.Errorf("failed to create player entity")
	}

	return id, nil
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
