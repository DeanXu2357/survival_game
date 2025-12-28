package domain

import (
	"fmt"

	"survival/internal/core/domain/state"
	"survival/internal/core/domain/system"
	"survival/internal/core/ports"
)

type Game struct {
	world     *state.World
	mapConfig *MapConfig

	movementSys system.MovementSystem
}

func NewGame(mapConfig *MapConfig) (*Game, error) {
	gridWidth := int(mapConfig.Dimensions.X / mapConfig.GridSize)
	gridHeight := int(mapConfig.Dimensions.Y / mapConfig.GridSize)

	g := &Game{
		world:       state.NewWorld(mapConfig.GridSize, gridWidth, gridHeight),
		mapConfig:   mapConfig,
		movementSys: *system.NewMovementSystem(),
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
		// TODO: refactor this , shouldn't set world property directly
		g.world.Collider.Add(id, state.Collider{
			Center:    state.Position{X: wallCfg.Center.X, Y: wallCfg.Center.Y},
			HalfSize:  wallCfg.HalfSize,
			ShapeType: state.ColliderBox,
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

func (g *Game) JoinPlayer() (state.EntityID, error) {
	spawnPoint := g.mapConfig.GetRandomSpawnPoint()
	if spawnPoint == nil {
		return 0, fmt.Errorf("no spawn point available")
	}

	id, ok := g.world.CreatePlayer(state.CreatePlayer{
		Position:      state.Position{X: spawnPoint.Position.X, Y: spawnPoint.Position.Y},
		Direction:     0,
		MovementSpeed: state.MovementSpeed(defaultPlayerMovementSpeed),
		RotationSpeed: state.RotationSpeed(defaultPlayerRotationSpeed),
		Radius:        defaultPlayerRadius,
		Health:        state.Health(defaultPlayerHealth),
	})
	if !ok {
		return 0, fmt.Errorf("failed to create player entity")
	}

	return id, nil
}

func (g *Game) UpdateInLoop(dt float64, playerInputs map[state.EntityID]ports.PlayerInput) {
	_ = g.movementSys.Update(dt, g.world, playerInputs)

	// visionDelta := g.visionSys.Update(dt, g.world, g.buf, positionDelta)
	// aliveDelta := g.combatSys.Update(dt, g.world, g.buf)

	g.world.ApplyCommands()
	// note: maybe can log here for debug ?
}

func (g *Game) Statics() []state.StaticEntity {
	return g.world.StaticEntities()
}

func (g *Game) PlayerSnapshotWithLocation(playerID state.EntityID) (state.PlayerSnapshotWithView, bool) {
	return g.world.PlayerSnapshotWithView(playerID)
}
