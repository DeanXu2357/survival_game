package engine

import (
	"fmt"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/system"
)

type Game struct {
	world     *state.World
	mapConfig *MapConfig
	systems   *state.SystemManager
}

func NewGame(mapConfig *MapConfig) (*Game, error) {
	gridWidth := int(mapConfig.Dimensions.X / mapConfig.GridSize)
	gridHeight := int(mapConfig.Dimensions.Y / mapConfig.GridSize)

	world := state.NewWorld(mapConfig.GridSize, gridWidth, gridHeight)
	world.Width = mapConfig.Dimensions.X
	world.Height = mapConfig.Dimensions.Y

	systems := state.NewSystemManager(world)
	systems.Register(system.NewBasicMovementSystem(world))

	g := &Game{
		world:     world,
		mapConfig: mapConfig,
		systems:   systems,
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

		collider := state.Collider{
			Center:    state.Position{X: wallCfg.Center.X, Y: wallCfg.Center.Y},
			HalfSize:  wallCfg.HalfSize,
			ShapeType: state.ColliderBox,
		}
		g.world.Collider.Upsert(id, collider)

		height := wallCfg.Height
		if height == 0 {
			height = state.DefaultWallHeight
		}
		vertBody := state.VerticalBody{
			BaseElevation: wallCfg.BaseElevation,
			Height:        height,
		}
		g.world.VerticalBody.Upsert(id, vertBody)

		g.world.EntityMeta.Upsert(id, state.WallMeta)

		min, max := collider.BoundingBox()
		g.world.Grid.Add(id, state.Bounds{
			MinX: min.X, MinY: min.Y,
			MaxX: max.X, MaxY: max.Y,
		}, state.LayerStatic)
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

	g.world.ApplyCommands()

	return id, nil
}

func (g *Game) Update(dt float64) {
	g.world.SyncInputBuffer()
	g.systems.Update(dt)
	g.world.ApplyCommands()
}

func (g *Game) SetPlayerInput(entityID state.EntityID, input ports.PlayerInput) {
	var mt state.MovementType
	if input.MovementType == ports.MovementTypeRelative {
		mt = state.MovementTypeRelative
	}

	g.world.SetInput(entityID, state.Input{
		MoveVertical:   input.MoveVertical,
		MoveHorizontal: input.MoveHorizontal,
		LookHorizontal: input.LookHorizontal,
		MovementType:   mt,
		Fire:           input.Fire,
		SwitchWeapon:   input.SwitchWeapon,
		Reload:         input.Reload,
		FastReload:     input.FastReload,
		Timestamp:      input.Timestamp,
	})
}

func (g *Game) Statics() []state.StaticEntity {
	return g.world.StaticEntities()
}

func (g *Game) PlayerSnapshotWithLocation(playerID state.EntityID) (state.PlayerSnapshotWithView, bool) {
	return g.world.PlayerSnapshotWithView(playerID)
}

func (g *Game) MapInfo() state.MapInfo {
	return g.world.MapInfo()
}
