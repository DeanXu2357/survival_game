package domain

import (
	"fmt"
	"math"
	"sync"

	"survival/internal/core/domain/vector"
)

// World is the ECS container that replaces the old State struct.
type World struct {
	// Entity Management
	entities *EntityManager

	// Component Managers (one per component type)
	positions       *ComponentManger[Position]
	velocities      *ComponentManger[Velocity]
	directions      *ComponentManger[Direction]
	circleColliders *ComponentManger[CircleCollider]
	boxColliders    *ComponentManger[BoxCollider]
	playerStats     *ComponentManger[PlayerStats]
	playerIdentity  *ComponentManger[PlayerIdentity]
	gridCells       *ComponentManger[GridCells]

	// Tag Component Managers (zero-size types for entity categorization)
	playerTags     *ComponentManger[PlayerTag]
	staticTags     *ComponentManger[StaticTag]
	projectileTags *ComponentManger[ProjectileTag]

	// Spatial Grid
	grid *Grid

	// Synchronization
	mu sync.RWMutex

	// Configuration
	allowNewPlayers bool
	playerCounter   int

	// Index mappings for network protocol compatibility
	playerIDToEntity map[string]EntityID
	entityToPlayerID map[EntityID]string

	// Map dimensions (for boundary checks)
	mapWidth  float64
	mapHeight float64
}

// NewWorld creates a new ECS world with the given grid parameters.
func NewWorld(gridCellSize float64, gridWidth, gridHeight int) *World {
	return &World{
		entities:         NewEntityManager(),
		positions:        NewComponentManager[Position](),
		velocities:       NewComponentManager[Velocity](),
		directions:       NewComponentManager[Direction](),
		circleColliders:  NewComponentManager[CircleCollider](),
		boxColliders:     NewComponentManager[BoxCollider](),
		playerStats:      NewComponentManager[PlayerStats](),
		playerIdentity:   NewComponentManager[PlayerIdentity](),
		gridCells:        NewComponentManager[GridCells](),
		playerTags:       NewComponentManager[PlayerTag](),
		staticTags:       NewComponentManager[StaticTag](),
		projectileTags:   NewComponentManager[ProjectileTag](),
		grid:             NewGrid(gridCellSize, gridWidth, gridHeight),
		allowNewPlayers:  true,
		playerCounter:    0,
		playerIDToEntity: make(map[string]EntityID),
		entityToPlayerID: make(map[EntityID]string),
		mapWidth:         float64(gridWidth) * gridCellSize,
		mapHeight:        float64(gridHeight) * gridCellSize,
	}
}

// NewWorldFromMap creates a world initialized with map configuration.
func NewWorldFromMap(mapConfig *MapConfig) *World {
	gridWidth := int(mapConfig.Dimensions.X/mapConfig.GridSize) + 1
	gridHeight := int(mapConfig.Dimensions.Y/mapConfig.GridSize) + 1

	world := &World{
		entities:         NewEntityManager(),
		positions:        NewComponentManager[Position](),
		velocities:       NewComponentManager[Velocity](),
		directions:       NewComponentManager[Direction](),
		circleColliders:  NewComponentManager[CircleCollider](),
		boxColliders:     NewComponentManager[BoxCollider](),
		playerStats:      NewComponentManager[PlayerStats](),
		playerIdentity:   NewComponentManager[PlayerIdentity](),
		gridCells:        NewComponentManager[GridCells](),
		playerTags:       NewComponentManager[PlayerTag](),
		staticTags:       NewComponentManager[StaticTag](),
		projectileTags:   NewComponentManager[ProjectileTag](),
		grid:             NewGrid(mapConfig.GridSize, gridWidth, gridHeight),
		allowNewPlayers:  true,
		playerCounter:    0,
		playerIDToEntity: make(map[string]EntityID),
		entityToPlayerID: make(map[EntityID]string),
		mapWidth:         mapConfig.Dimensions.X,
		mapHeight:        mapConfig.Dimensions.Y,
	}

	// Create wall entities from map configuration
	for _, wallConfig := range mapConfig.Walls {
		world.CreateWall(wallConfig.ID, wallConfig.Center, wallConfig.HalfSize, wallConfig.Rotation)
	}

	return world
}

// generatePlayerID generates a unique legacy player ID for network protocol.
func (w *World) generatePlayerID() string {
	w.playerCounter++
	return fmt.Sprintf("player-%d", w.playerCounter)
}

// CreatePlayer creates a new player entity and returns the EntityID and legacy playerID.
func (w *World) CreatePlayer(sessionID string, position vector.Vector2D) (EntityID, string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.allowNewPlayers {
		return 0, "", fmt.Errorf("adding new players is not allowed")
	}

	// Allocate entity
	entityID, ok := w.entities.Alloc()
	if !ok {
		return 0, "", fmt.Errorf("failed to allocate entity: max entities reached")
	}

	playerID := w.generatePlayerID()

	// Add components
	w.positions.Add(entityID, Position{X: position.X, Y: position.Y})
	w.velocities.Add(entityID, Velocity{X: 0, Y: 0})
	w.directions.Add(entityID, Direction{Angle: 0})
	w.circleColliders.Add(entityID, CircleCollider{Radius: 0.5})
	w.playerStats.Add(entityID, PlayerStats{
		Health:        100,
		IsAlive:       true,
		MovementSpeed: playerBaseMovementSpeed,
		RotationSpeed: playerBaseRotationSpeed,
	})
	w.playerIdentity.Add(entityID, PlayerIdentity{
		PlayerID:  playerID,
		SessionID: sessionID,
	})
	w.playerTags.Add(entityID, PlayerTag{})

	// Add to spatial grid
	radius := 0.5
	bounds := Bounds{
		MinX: position.X - radius,
		MinY: position.Y - radius,
		MaxX: position.X + radius,
		MaxY: position.Y + radius,
	}
	indexes := w.grid.Add(entityID, bounds, LayerPlayer)
	w.gridCells.Add(entityID, GridCells{Indexes: indexes})

	// Update mappings for network protocol compatibility
	w.playerIDToEntity[playerID] = entityID
	w.entityToPlayerID[entityID] = playerID

	return entityID, playerID, nil
}

// RemovePlayer removes a player entity and all its components.
func (w *World) RemovePlayer(entityID EntityID) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Get player ID for cleanup
	if playerID, ok := w.entityToPlayerID[entityID]; ok {
		delete(w.playerIDToEntity, playerID)
		delete(w.entityToPlayerID, entityID)
	}

	// Remove from grid
	if gc := w.gridCells.Get(entityID); gc != nil {
		indexes := make([]uint64, len(gc.Indexes))
		for i, idx := range gc.Indexes {
			indexes[i] = uint64(idx)
		}
		w.grid.Remove(indexes, entityID)
	}

	// Remove all components
	w.positions.Remove(entityID)
	w.velocities.Remove(entityID)
	w.directions.Remove(entityID)
	w.circleColliders.Remove(entityID)
	w.playerStats.Remove(entityID)
	w.playerIdentity.Remove(entityID)
	w.gridCells.Remove(entityID)
	w.playerTags.Remove(entityID)

	// Free entity
	w.entities.Free(entityID)
}

// CreateWall creates a static wall entity.
func (w *World) CreateWall(id string, center, halfSize vector.Vector2D, rotation float64) EntityID {
	entityID, ok := w.entities.Alloc()
	if !ok {
		panic("failed to allocate wall entity: max entities reached")
	}

	// Add components
	w.positions.Add(entityID, Position{X: center.X, Y: center.Y})
	w.boxColliders.Add(entityID, BoxCollider{
		HalfWidth:  halfSize.X,
		HalfHeight: halfSize.Y,
		Rotation:   rotation,
	})
	w.staticTags.Add(entityID, StaticTag{})

	// Calculate AABB bounds for grid
	bounds := calculateAABB(center, halfSize, rotation)
	indexes := w.grid.Add(entityID, bounds, LayerStatic)
	w.gridCells.Add(entityID, GridCells{Indexes: indexes})

	return entityID
}

// calculateAABB computes axis-aligned bounding box for a potentially rotated rectangle.
func calculateAABB(center, halfSize vector.Vector2D, rotation float64) Bounds {
	cos := math.Cos(rotation)
	sin := math.Sin(rotation)

	extentX := math.Abs(halfSize.X*cos) + math.Abs(halfSize.Y*sin)
	extentY := math.Abs(halfSize.X*sin) + math.Abs(halfSize.Y*cos)

	return Bounds{
		MinX: center.X - extentX,
		MinY: center.Y - extentY,
		MaxX: center.X + extentX,
		MaxY: center.Y + extentY,
	}
}

// GetEntityByPlayerID looks up an entity by legacy player ID.
func (w *World) GetEntityByPlayerID(playerID string) (EntityID, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	entityID, ok := w.playerIDToEntity[playerID]
	return entityID, ok
}

// GetPlayerID gets the legacy player ID for an entity.
func (w *World) GetPlayerID(entityID EntityID) (string, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	playerID, ok := w.entityToPlayerID[entityID]
	return playerID, ok
}

// ForEachPlayer iterates over all player entities and calls the provided function.
func (w *World) ForEachPlayer(fn func(entityID EntityID)) {
	for _, entityID := range w.playerTags.IndexToEntityID {
		if w.entities.IsAlive(entityID) {
			fn(entityID)
		}
	}
}

// GetNearbyEntities finds entities near a position within radius that match the layer mask.
func (w *World) GetNearbyEntities(pos Position, radius float64, layerMask LayerMask) []EntityID {
	bounds := Bounds{
		MinX: pos.X - radius,
		MinY: pos.Y - radius,
		MaxX: pos.X + radius,
		MaxY: pos.Y + radius,
	}

	seen := make(map[EntityID]bool)
	var result []EntityID

	for _, cell := range w.grid.CellsInBounds(bounds) {
		for _, entry := range cell.entries {
			if !seen[entry.EntityID] && entry.Layer.Has(layerMask) {
				seen[entry.EntityID] = true
				result = append(result, entry.EntityID)
			}
		}
	}

	return result
}

// PlayerCount returns the number of active players.
func (w *World) PlayerCount() int {
	return len(w.playerTags.IndexToEntityID)
}

// SetAllowNewPlayers sets whether new players can join.
func (w *World) SetAllowNewPlayers(allow bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.allowNewPlayers = allow
}

// Positions returns the positions component manager for external access.
func (w *World) Positions() *ComponentManger[Position] {
	return w.positions
}

// Directions returns the directions component manager for external access.
func (w *World) Directions() *ComponentManger[Direction] {
	return w.directions
}

// PlayerStats returns the player stats component manager for external access.
func (w *World) PlayerStatsManager() *ComponentManger[PlayerStats] {
	return w.playerStats
}

// CircleColliders returns the circle colliders component manager for external access.
func (w *World) CircleColliders() *ComponentManger[CircleCollider] {
	return w.circleColliders
}

// BoxColliders returns the box colliders component manager for external access.
func (w *World) BoxColliders() *ComponentManger[BoxCollider] {
	return w.boxColliders
}

// Grid returns the spatial grid for external access.
func (w *World) Grid() *Grid {
	return w.grid
}

// UpdateEntityGridPosition updates an entity's position in the spatial grid.
func (w *World) UpdateEntityGridPosition(entityID EntityID, newPos Position, radius float64, layer LayerMask) {
	// Remove from old cells
	gc := w.gridCells.Get(entityID)
	if gc != nil {
		indexes := make([]uint64, len(gc.Indexes))
		for i, idx := range gc.Indexes {
			indexes[i] = uint64(idx)
		}
		w.grid.Remove(indexes, entityID)
	}

	// Add to new cells
	bounds := Bounds{
		MinX: newPos.X - radius,
		MinY: newPos.Y - radius,
		MaxX: newPos.X + radius,
		MaxY: newPos.Y + radius,
	}
	newIndexes := w.grid.Add(entityID, bounds, layer)

	// Update component
	if gc != nil {
		gc.Indexes = newIndexes
	} else {
		w.gridCells.Add(entityID, GridCells{Indexes: newIndexes})
	}
}

// ToClientState converts ECS world state to the legacy ClientGameState format.
// This maintains backward compatibility with the existing WebSocket protocol.
func (w *World) ToClientState() *ClientGameState {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Build players map using legacy Player struct format
	players := make(map[string]*Player)

	for _, entityID := range w.playerTags.IndexToEntityID {
		if !w.entities.IsAlive(entityID) {
			continue
		}

		pos := w.positions.Get(entityID)
		dir := w.directions.Get(entityID)
		stats := w.playerStats.Get(entityID)
		identity := w.playerIdentity.Get(entityID)
		collider := w.circleColliders.Get(entityID)

		if pos == nil || identity == nil {
			continue
		}

		player := &Player{
			ID:       identity.PlayerID,
			Position: pos.ToVector2D(),
		}

		if dir != nil {
			player.Direction = dir.Angle
		}
		if stats != nil {
			player.Health = stats.Health
			player.IsAlive = stats.IsAlive
			player.MovementSpeed = stats.MovementSpeed
			player.RotationSpeed = stats.RotationSpeed
		}
		if collider != nil {
			player.Radius = collider.Radius
		}

		players[identity.PlayerID] = player
	}

	// Build walls DTOs
	var walls []*WallDTO

	for _, entityID := range w.staticTags.IndexToEntityID {
		if !w.entities.IsAlive(entityID) {
			continue
		}

		pos := w.positions.Get(entityID)
		box := w.boxColliders.Get(entityID)

		if pos == nil || box == nil {
			continue
		}

		wallDTO := &WallDTO{
			ID:       fmt.Sprintf("wall-%d", entityID),
			Center:   pos.ToVector2D(),
			HalfSize: vector.Vector2D{X: box.HalfWidth, Y: box.HalfHeight},
			Rotation: box.Rotation,
		}
		walls = append(walls, wallDTO)
	}

	return &ClientGameState{
		Players:     players,
		Walls:       walls,
		Projectiles: nil, // TODO: implement projectiles
		Timestamp:   0,
	}
}

// ToStaticData converts world to static data for initial client load.
func (w *World) ToStaticData() *StaticGameData {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var walls []*WallDTO

	for _, entityID := range w.staticTags.IndexToEntityID {
		if !w.entities.IsAlive(entityID) {
			continue
		}

		pos := w.positions.Get(entityID)
		box := w.boxColliders.Get(entityID)

		if pos == nil || box == nil {
			continue
		}

		wallDTO := &WallDTO{
			ID:       fmt.Sprintf("wall-%d", entityID),
			Center:   pos.ToVector2D(),
			HalfSize: vector.Vector2D{X: box.HalfWidth, Y: box.HalfHeight},
			Rotation: box.Rotation,
		}
		walls = append(walls, wallDTO)
	}

	return &StaticGameData{
		Type:      "staticData",
		Walls:     walls,
		MapWidth:  int(w.mapWidth),
		MapHeight: int(w.mapHeight),
	}
}
