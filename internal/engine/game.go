package engine

import (
	"fmt"
	"time"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/system"
)

type Game struct {
	world     *state.World
	mapConfig *MapConfig
	systems   *state.SystemManager

	// Tick-based time tracking
	currentTick   uint64    // Current game tick (increments each Update call)
	startTime     time.Time // Wall-clock time when game loop started
	lastUpdate    time.Time // Wall-clock time of last Update() call
	isInitialized bool      // Whether StartGameLoop() has been called
}

func NewGame(mapConfig *MapConfig) (*Game, error) {
	gridWidth := int(mapConfig.Dimensions.X / mapConfig.GridSize)
	gridHeight := int(mapConfig.Dimensions.Y / mapConfig.GridSize)

	world := state.NewWorld(mapConfig.GridSize, gridWidth, gridHeight)
	world.Width = mapConfig.Dimensions.X
	world.Height = mapConfig.Dimensions.Y

	systems := state.NewSystemManager(world)

	g := &Game{
		world:     world,
		mapConfig: mapConfig,
		systems:   systems,
	}

	systems.Register(system.NewWeaponSwitchSystem(world, &g.currentTick))
	systems.Register(system.NewBasicMovementSystem(world))
	systems.Register(system.NewProjectileSystem(world, &g.currentTick))
	systems.Register(system.NewWeaponFireSystem(world, &g.currentTick))

	if err := g.loadMapEntities(mapConfig); err != nil {
		return nil, err
	}
	return g, nil
}

// StartGameLoop initializes the internal tick counter and game clock.
// This must be called before the first Update() call.
// Calling it multiple times will reset the tick counter and game time.
func (g *Game) StartGameLoop() {
	now := time.Now()
	g.startTime = now
	g.lastUpdate = now
	g.currentTick = 0
	g.isInitialized = true
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

		minP, maxP := collider.BoundingBox()
		g.world.Grid.Add(id, state.Bounds{
			MinX: minP.X, MinY: minP.Y,
			MaxX: maxP.X, MaxY: maxP.Y,
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
	// Auto-initialize if StartGameLoop wasn't called
	if !g.isInitialized {
		g.StartGameLoop()
	}

	// Clamp dt to prevent physics instability from lag spikes
	if dt > ports.MaxFrameTime {
		dt = ports.MaxFrameTime
	}

	// Increment tick counter (deterministic, always +1 per update)
	g.currentTick++
	g.lastUpdate = time.Now()

	// Run game systems with clamped dt
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

// CurrentTick returns the current game tick count.
// Each tick represents one Update() call at the target tick rate (60 FPS).
func (g *Game) CurrentTick() uint64 {
	return g.currentTick
}

// ElapsedSeconds returns the total simulated game time in seconds.
// Calculated from tick count: seconds = ticks / TickRate
func (g *Game) ElapsedSeconds() float64 {
	return float64(g.currentTick) / ports.TargetTickRate
}

// StartTime returns the wall-clock time when the game loop was started.
func (g *Game) StartTime() time.Time {
	return g.startTime
}

// LastUpdateTime returns the wall-clock time of the most recent Update() call.
func (g *Game) LastUpdateTime() time.Time {
	return g.lastUpdate
}

// IsInitialized returns whether StartGameLoop() has been called.
func (g *Game) IsInitialized() bool {
	return g.isInitialized
}
