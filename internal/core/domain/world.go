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
