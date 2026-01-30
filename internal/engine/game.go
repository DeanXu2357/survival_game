package engine

import (
	"fmt"

	"survival/internal/engine/ports"
	state2 "survival/internal/engine/state"
	"survival/internal/engine/system"
)

type Game struct {
	world     *state2.World
	mapConfig *MapConfig

	movementSys system.BasicMovementSystem
}

func NewGame(mapConfig *MapConfig) (*Game, error) {
	gridWidth := int(mapConfig.Dimensions.X / mapConfig.GridSize)
	gridHeight := int(mapConfig.Dimensions.Y / mapConfig.GridSize)

	world := state2.NewWorld(mapConfig.GridSize, gridWidth, gridHeight)
	world.Width = mapConfig.Dimensions.X
	world.Height = mapConfig.Dimensions.Y

	g := &Game{
		world:       world,
		mapConfig:   mapConfig,
		movementSys: *system.NewBasicMovementSystem(),
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

		collider := state2.Collider{
			Center:    state2.Position{X: wallCfg.Center.X, Y: wallCfg.Center.Y},
			HalfSize:  wallCfg.HalfSize,
			ShapeType: state2.ColliderBox,
		}
		g.world.Collider.Upsert(id, collider)

		height := wallCfg.Height
		if height == 0 {
			height = state2.DefaultWallHeight
		}
		vertBody := state2.VerticalBody{
			BaseElevation: wallCfg.BaseElevation,
			Height:        height,
		}
		g.world.VerticalBody.Upsert(id, vertBody)

		g.world.EntityMeta.Upsert(id, state2.WallMeta)

		min, max := collider.BoundingBox()
		g.world.Grid.Add(id, state2.Bounds{
			MinX: min.X, MinY: min.Y,
			MaxX: max.X, MaxY: max.Y,
		}, state2.LayerStatic)
	}
	return nil
}

const (
	defaultPlayerMovementSpeed float64 = 5
	defaultPlayerRotationSpeed float64 = 2
	defaultPlayerRadius        float64 = 0.5
	defaultPlayerHealth        int     = 100
)

func (g *Game) JoinPlayer() (state2.EntityID, error) {
	spawnPoint := g.mapConfig.GetRandomSpawnPoint()
	if spawnPoint == nil {
		return 0, fmt.Errorf("no spawn point available")
	}

	id, ok := g.world.CreatePlayer(state2.CreatePlayer{
		Position:      state2.Position{X: spawnPoint.Position.X, Y: spawnPoint.Position.Y},
		Direction:     0,
		MovementSpeed: state2.MovementSpeed(defaultPlayerMovementSpeed),
		RotationSpeed: state2.RotationSpeed(defaultPlayerRotationSpeed),
		Radius:        defaultPlayerRadius,
		Health:        state2.Health(defaultPlayerHealth),
	})
	if !ok {
		return 0, fmt.Errorf("failed to create player entity")
	}

	g.world.ApplyCommands()

	return id, nil
}

func (g *Game) UpdateInLoop(dt float64, playerInputs map[state2.EntityID]ports.PlayerInput) {
	_ = g.movementSys.Update(dt, g.world, transformPlayerInputs(playerInputs))

	// visionDelta := g.visionSys.Update(dt, g.world, g.buf, positionDelta)
	// aliveDelta := g.combatSys.Update(dt, g.world, g.buf)

	g.world.ApplyCommands()
	// note: maybe can log here for debug ?
}

func transformPlayerInputs(inputs map[state2.EntityID]ports.PlayerInput) map[state2.EntityID]system.PlayerInput {
	result := make(map[state2.EntityID]system.PlayerInput)
	for id, input := range inputs {
		var mt system.MovementType
		if input.MovementType == ports.MovementTypeRelative {
			mt = system.MovementTypeRelative
		}
		result[id] = system.PlayerInput{
			MoveVertical:   input.MoveVertical,
			MoveHorizontal: input.MoveHorizontal,
			LookHorizontal: input.LookHorizontal,
			MovementType:   mt,
			SwitchWeapon:   input.SwitchWeapon,
			Reload:         input.Reload,
			FastReload:     input.FastReload,
			Fire:           input.Fire,
			Timestamp:      input.Timestamp,
		}
	}
	return result
}

func (g *Game) Statics() []state2.StaticEntity {
	return g.world.StaticEntities()
}

func (g *Game) PlayerSnapshotWithLocation(playerID state2.EntityID) (state2.PlayerSnapshotWithView, bool) {
	return g.world.PlayerSnapshotWithView(playerID)
}

func (g *Game) MapInfo() state2.MapInfo {
	return g.world.MapInfo()
}
