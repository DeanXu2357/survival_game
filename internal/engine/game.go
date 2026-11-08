package engine

import (
	"fmt"
	"math"
	"time"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/system"
	"survival/internal/engine/system/ai"
	goapturret "survival/internal/engine/system/ai/goap/turret"
	"survival/internal/engine/system/ai/turret"
)

type Game struct {
	world     *state.World
	mapConfig *MapConfig
	systems   *state.SystemManager

	fistDefID   state.EntityID // EntityID for the Fist weapon definition
	turretDefID state.EntityID // EntityID for the Turret weapon definition

	enemyAI *system.EnemyAISystem

	// Tick-based time tracking
	currentTick   ports.Tick // Current game tick (increments each Update call)
	startTime     time.Time  // Wall-clock time when game loop started
	lastUpdate    time.Time  // Wall-clock time of last Update() call
	isInitialized bool       // Whether StartGameLoop() has been called
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

	aiRegistry := map[string]ai.Behavior{
		turret.StationaryTurretKey: turret.NewStationaryTurretBehavior(),
		goapturret.GOAPTurretKey:   goapturret.NewBehavior(),
	}
	g.enemyAI = system.NewEnemyAISystem(world, &g.currentTick, aiRegistry)

	systems.Register(system.NewInventorySystem(world, &g.currentTick))
	systems.Register(system.NewBasicMovementSystem(world))
	systems.Register(system.NewProjectileSystem(world, &g.currentTick))
	systems.Register(g.enemyAI)
	systems.Register(system.NewWeaponFireSystem(world, &g.currentTick))
	systems.Register(system.NewReviveSystem(world, &g.currentTick))

	// Reserve entity 0 so no real entity gets EntityID(0),
	// which is used as the zero-value sentinel in IsEmpty() checks.
	world.Entity.Alloc()

	// Create the Fist weapon definition entity
	fistDefID, ok := world.CreateItemDefEntity(state.ItemConfig{
		Name:         "Fist",
		Type:         state.ItemTypeWeapon,
		MaxStack:     1,
		WeaponConfig: state.FistSpec,
	})
	if !ok {
		return nil, fmt.Errorf("failed to create fist item definition")
	}
	g.fistDefID = fistDefID

	turretDefID, ok := world.CreateItemDefEntity(state.ItemConfig{
		Name:         "Turret",
		Type:         state.ItemTypeWeapon,
		MaxStack:     1,
		WeaponConfig: state.TurretSpec,
	})
	if !ok {
		return nil, fmt.Errorf("failed to create turret item definition")
	}
	g.turretDefID = turretDefID

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

		height := wallCfg.Height
		if height == 0 {
			height = state.DefaultWallHeight
		}
		collider := state.Collider{
			Center:        state.Position{X: wallCfg.Center.X, Y: wallCfg.Center.Y},
			HalfSize:      wallCfg.HalfSize,
			ShapeType:     state.ColliderBox,
			BaseElevation: wallCfg.BaseElevation,
			Height:        height,
		}
		g.world.Collider.Upsert(id, collider)

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

const (
	defaultDummySpawnOffset    float64 = 8.0
	defaultDummyRespawnSeconds float64 = 3.0
)

const (
	defaultTurretSpawnOffset    float64 = 12.0
	defaultTurretRespawnSeconds float64 = 5.0
	defaultTurretRadius         float64 = 0.6
	defaultTurretHealth         int     = 100
)

// SpawnTrainingDummyNearSpawn spawns a stationary, auto-respawning training
// dummy a short distance in front of the map's primary spawn point. The
// dummy is modeled as a zero-speed player with an attached Revive component.
// Intended for single-player practice; safe to skip for tests or multiplayer.
func (g *Game) SpawnTrainingDummyNearSpawn() (state.EntityID, error) {
	spawn := g.mapConfig.GetRandomSpawnPoint()
	if spawn == nil {
		return 0, fmt.Errorf("no spawn point available")
	}

	pos := state.Position{
		X: spawn.Position.X,
		Y: spawn.Position.Y - defaultDummySpawnOffset,
	}
	hp := state.Health(defaultPlayerHealth)

	id, ok := g.world.CreatePlayer(state.CreatePlayer{
		Position:      pos,
		Direction:     0,
		MovementSpeed: 0,
		RotationSpeed: 0,
		Radius:        defaultPlayerRadius,
		Health:        hp,
		FistDefID:     g.fistDefID,
	})
	if !ok {
		return 0, fmt.Errorf("failed to allocate dummy entity")
	}

	g.world.UpdatePlayer(id, state.UpdatePlayer{
		UpdateMeta: state.ComponentMeta | state.ComponentRevive,
		Meta:       state.PlayerMeta | state.ComponentRevive,
		Revive: state.Revive{
			SpawnHealth:       hp,
			RespawnDelayTicks: ports.TicksFromSeconds(defaultDummyRespawnSeconds),
		},
	})

	g.world.ApplyCommands()
	return id, nil
}

// SpawnTrainingTurretNearSpawn spawns a stationary, auto-respawning turret
// driven by the legacy state-machine AI.
func (g *Game) SpawnTrainingTurretNearSpawn() (state.EntityID, error) {
	return g.spawnTurretNearSpawn(turret.StationaryTurretKey)
}

// SpawnGOAPTurretNearSpawn spawns the same training turret but driven by the
// GOAP-based AI behavior.
func (g *Game) SpawnGOAPTurretNearSpawn() (state.EntityID, error) {
	return g.spawnTurretNearSpawn(goapturret.GOAPTurretKey)
}

// spawnTurretNearSpawn spawns a stationary, auto-respawning turret a short
// distance from the map's primary spawn point, facing the spawn so its frontal
// cone covers anyone who walks toward it, and assigns the given AI behavior key.
// Intended for single-player training; safe to skip for tests or multiplayer.
func (g *Game) spawnTurretNearSpawn(aiKey string) (state.EntityID, error) {
	spawn := g.mapConfig.GetRandomSpawnPoint()
	if spawn == nil {
		return 0, fmt.Errorf("no spawn point available")
	}

	pos := state.Position{
		X: spawn.Position.X,
		Y: spawn.Position.Y + defaultTurretSpawnOffset,
	}
	// Face back toward the spawn point (negative Y direction). With the
	// Forward convention (0 = north/up), Direction = math.Pi points down.
	const facingSpawn state.Direction = state.Direction(math.Pi)

	hp := state.Health(defaultTurretHealth)
	id, ok := g.world.CreatePlayer(state.CreatePlayer{
		Position:      pos,
		Direction:     facingSpawn,
		MovementSpeed: 0,
		RotationSpeed: 0,
		Radius:        defaultTurretRadius,
		Health:        hp,
		FistDefID:     g.fistDefID,
	})
	if !ok {
		return 0, fmt.Errorf("failed to allocate turret entity")
	}

	inv := state.DefaultInventory(g.fistDefID, ports.ItemSlotCount)
	inv.Weapons[2] = state.WeaponSlot{ItemID: g.turretDefID}
	inv.CurrentWeaponIndex = 2

	g.world.UpdatePlayer(id, state.UpdatePlayer{
		UpdateMeta: state.ComponentMeta | state.ComponentRevive | state.ComponentInventory,
		Meta:       state.PlayerMeta | state.ComponentRevive,
		Revive: state.Revive{
			SpawnHealth:       hp,
			RespawnDelayTicks: ports.TicksFromSeconds(defaultTurretRespawnSeconds),
		},
		Inventory: inv,
	})

	g.world.ApplyCommands()
	g.world.AI.Upsert(id, state.AIBehavior{Key: aiKey})
	return id, nil
}

func (g *Game) ThreatIndicators() []state.ThreatIndicator {
	return g.enemyAI.Indicators()
}

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
		FistDefID:     g.fistDefID,
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
		PickupEntityID: input.PickupEntityID,
		DropSlotIndex:  input.DropSlotIndex,
		Timestamp:      input.Timestamp,
	})
}

// WallEntities returns renderable entities that are not damageable (walls).
func (g *Game) WallEntities() []state.StaticEntity {
	all := g.world.StaticEntities()
	walls := make([]state.StaticEntity, 0, len(all))
	for _, entity := range all {
		if _, hasHealth := g.world.Health.Get(entity.ID); hasHealth {
			continue
		}
		walls = append(walls, entity)
	}
	return walls
}

// PlayerColliders returns renderable entities that can take damage (players, dummies), excluding the viewer.
func (g *Game) PlayerColliders(exclude state.EntityID) []state.StaticEntity {
	all := g.world.StaticEntities()
	players := make([]state.StaticEntity, 0, len(all))
	for _, entity := range all {
		if entity.ID == exclude {
			continue
		}
		if _, hasHealth := g.world.Health.Get(entity.ID); !hasHealth {
			continue
		}
		players = append(players, entity)
	}
	return players
}

func (g *Game) PlayerSnapshotWithLocation(playerID state.EntityID) (state.PlayerSnapshotWithView, bool) {
	return g.world.PlayerSnapshotWithView(playerID)
}

func (g *Game) MapInfo() state.MapInfo {
	return g.world.MapInfo()
}

// CurrentTick returns the current game tick count.
// Each tick represents one Update() call at the target tick rate (60 FPS).
func (g *Game) CurrentTick() ports.Tick {
	return g.currentTick
}

// ElapsedSeconds returns the total simulated game time in seconds.
// Calculated from tick count: seconds = ticks / TickRate
func (g *Game) ElapsedSeconds() float64 {
	return g.currentTick.Seconds()
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

// PlayerHealth returns the health of the given player.
func (g *Game) PlayerHealth(playerID state.EntityID) (state.Health, bool) {
	return g.world.Health.Get(playerID)
}

// PlayerInventory returns the inventory of the given player.
func (g *Game) PlayerInventory(playerID state.EntityID) (state.Inventory, bool) {
	return g.world.Inventory.Get(playerID)
}

// ProjectileCount returns the number of active projectile entities.
func (g *Game) ProjectileCount() int {
	n := 0
	for range g.world.Projectile.All() {
		n++
	}
	return n
}

// GroundItemSnapshots returns snapshots of all ground item entities.
func (g *Game) GroundItemSnapshots() []state.GroundItemSnapshot {
	var items []state.GroundItemSnapshot
	for entityID, gi := range g.world.GroundItem.All() {
		pos, ok := g.world.Position.Get(entityID)
		if !ok {
			continue
		}
		items = append(items, state.GroundItemSnapshot{
			ID:         entityID,
			Position:   pos,
			GroundItem: gi,
		})
	}
	return items
}

// RegisterItemDef creates a new item definition entity.
func (g *Game) RegisterItemDef(config state.ItemConfig) (state.EntityID, error) {
	id, ok := g.world.CreateItemDefEntity(config)
	if !ok {
		return 0, fmt.Errorf("failed to create item definition")
	}
	return id, nil
}

// SpawnGroundItem creates a ground item entity at the given position.
func (g *Game) SpawnGroundItem(pos state.Position, itemDefID state.EntityID, quantity, ammo int) (state.EntityID, error) {
	id, ok := g.world.CreateGroundItemEntity(state.CreateGroundItem{
		Position:  pos,
		ItemDefID: itemDefID,
		Quantity:  quantity,
		Ammo:      ammo,
	})
	if !ok {
		return 0, fmt.Errorf("failed to create ground item")
	}
	g.world.ApplyCommands()
	return id, nil
}

// SetPlayerInventory overwrites the inventory of the given player.
func (g *Game) SetPlayerInventory(playerID state.EntityID, inv state.Inventory) error {
	if !g.world.Entity.IsAlive(playerID) {
		return fmt.Errorf("player %d not alive", playerID)
	}
	g.world.UpdatePlayer(playerID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory,
		Inventory:  inv,
	})
	g.world.ApplyCommands()
	return nil
}
