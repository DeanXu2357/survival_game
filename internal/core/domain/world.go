package domain

type World struct {
	Entity *EntityManager

	EntityMeta    ComponentManger[Meta]
	Position      ComponentManger[Position]
	Direction     ComponentManger[Direction]
	MovementSpeed ComponentManger[MovementSpeed]
	RotationSpeed ComponentManger[RotationSpeed]

	PlayerShape ComponentManger[PlayerShape]
	Health      ComponentManger[Health]

	WallShape ComponentManger[WallShape]

	Grid Grid
}

func NewWorld(gridCellSize float64, gridWidth, gridHeight int) *World {
	return &World{
		Entity:        NewEntityManager(),
		EntityMeta:    *NewComponentManager[Meta](),
		Position:      *NewComponentManager[Position](),
		Direction:     *NewComponentManager[Direction](),
		MovementSpeed: *NewComponentManager[MovementSpeed](),
		RotationSpeed: *NewComponentManager[RotationSpeed](),
		PlayerShape:   *NewComponentManager[PlayerShape](),
		Health:        *NewComponentManager[Health](),
		WallShape:     *NewComponentManager[WallShape](),
		Grid:          *NewGrid(gridCellSize, gridWidth, gridHeight),
	}
}

// CreateEntity allocates a new entity and returns its ID.
// Should not be called directly, use command buffer function instead.
func (w *World) CreateEntity() (EntityID, bool) {
	return w.Entity.Alloc()
}

// DestroyEntity removes an entity and all its associated components.
// Should not be called directly, use command buffer function instead.
func (w *World) DestroyEntity(e EntityID) bool {
	// todo: lock up meta components and remove all components associated with entity e
	return w.Entity.Free(e)
}

type PlayerConfig struct {
	Position      Position
	Direction     Direction
	MovementSpeed MovementSpeed
	RotationSpeed RotationSpeed
	Radius        float64
	Health        Health
}

// CreatePlayer allocates a player entity and adds all player components.
// This bypasses CommandBuffer for immediate effect since EntityID must be returned synchronously.
func (w *World) CreatePlayer(cfg PlayerConfig) (EntityID, bool) {
	id, ok := w.Entity.Alloc()
	if !ok {
		return 0, false
	}

	w.Position.Add(id, cfg.Position)
	w.Direction.Add(id, cfg.Direction)
	w.MovementSpeed.Add(id, cfg.MovementSpeed)
	w.RotationSpeed.Add(id, cfg.RotationSpeed)
	w.PlayerShape.Add(id, PlayerShape{Center: &cfg.Position, Radius: cfg.Radius})
	w.Health.Add(id, cfg.Health)

	meta := ComponentPosition.Set(ComponentDirection).Set(ComponentMovementSpeed).Set(ComponentRotationSpeed).Set(ComponentPlayerShape).Set(ComponentHealth)
	w.EntityMeta.Add(id, meta)

	return id, true
}
