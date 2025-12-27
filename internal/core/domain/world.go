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

	buf *CommandBuffer
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
		buf:           NewCommandBuffer(),
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

// CreatePlayer allocates a player entity and adds all player components.
// This bypasses CommandBuffer for immediate effect since EntityID must be returned synchronously.
func (w *World) CreatePlayer(cfg CreatePlayer) (EntityID, bool) {
	id, ok := w.Entity.Alloc()
	if !ok {
		return 0, false
	}

	w.UpdatePlayer(
		id,
		UpdatePlayer{
			UpdateMeta:    PlayerMeta,
			Position:      cfg.Position,
			Direction:     cfg.Direction,
			MovementSpeed: cfg.MovementSpeed,
			RotationSpeed: cfg.RotationSpeed,
			Meta:          PlayerMeta,
			PlayerShape:   PlayerShape{cfg.Position, cfg.Radius},
			Health:        cfg.Health,
		},
	)

	return id, true
}

type CreatePlayer struct {
	Position      Position
	Direction     Direction
	MovementSpeed MovementSpeed
	RotationSpeed RotationSpeed
	Radius        float64
	Health        Health
}

func (w *World) UpdatePlayer(id EntityID, player UpdatePlayer) {
	w.buf.Push(WorldCommand{
		EntityID:      id,
		UpdateMeta:    player.UpdateMeta,
		Position:      player.Position,
		Direction:     player.Direction,
		Meta:          player.Meta,
		RotationSpeed: player.RotationSpeed,
		MovementSpeed: player.MovementSpeed,
		PlayerShape:   player.PlayerShape,
		Health:        player.Health,
	})
}

type UpdatePlayer struct {
	UpdateMeta Meta
	Position
	Direction
	MovementSpeed
	RotationSpeed
	Meta
	PlayerShape
	Health
}

func (w *World) ApplyCommands() {
	// todo: implement command application logic
	panic("not implemented")
}
