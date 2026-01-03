package state

type World struct {
	Entity *EntityManager

	EntityMeta    ComponentManager[Meta]
	Position      ComponentManager[Position]
	Direction     ComponentManager[Direction]
	MovementSpeed ComponentManager[MovementSpeed]
	RotationSpeed ComponentManager[RotationSpeed]
	ViewIDs       ComponentManager[ViewIDs]

	PlayerHitbox ComponentManager[PlayerHitbox]
	Health       ComponentManager[Health]

	Collider ComponentManager[Collider]

	Grid Grid

	buf *CommandBuffer

	Width, Height float64
}

func NewWorld(gridCellSize float64, gridWidth, gridHeight int) *World {
	return &World{
		Entity:        NewEntityManager(),
		EntityMeta:    *NewComponentManager[Meta](), // TODO: refactor this, use pointer or not
		Position:      *NewComponentManager[Position](),
		Direction:     *NewComponentManager[Direction](),
		MovementSpeed: *NewComponentManager[MovementSpeed](),
		RotationSpeed: *NewComponentManager[RotationSpeed](),
		ViewIDs:       *NewComponentManager[ViewIDs](),
		PlayerHitbox:  *NewComponentManager[PlayerHitbox](),
		Health:        *NewComponentManager[Health](),
		Collider:      *NewComponentManager[Collider](),
		Grid:          *NewGrid(gridCellSize, gridWidth, gridHeight),
		buf:           NewCommandBuffer(),
		Width:         0,
		Height:        0,
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
			PlayerHitbox:  PlayerHitbox{cfg.Position, cfg.Radius},
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
		PlayerShape:   player.PlayerHitbox,
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
	PlayerHitbox
	Health
}

func (w *World) ApplyCommands() {
	for !w.buf.IsEmpty() {
		cmd, ok := w.buf.Pop()
		if !ok {
			// TODO: log error
			continue
		}

		entityID := cmd.EntityID
		if !w.Entity.IsAlive(entityID) {
			continue
		}

		meta, exist := w.EntityMeta.Get(entityID)
		if !exist {
			continue
		}

		if meta.Has(ComponentPosition) {
			if !w.Position.Set(entityID, cmd.Position) {
				// TODO: log error
			}
		}
		if meta.Has(ComponentDirection) {
			if !w.Direction.Set(entityID, cmd.Direction) {
				// TODO: log error
			}
		}
		if meta.Has(ComponentMovementSpeed) {
			if !w.MovementSpeed.Set(entityID, cmd.MovementSpeed) {
				// TODO: log error
			}
		}
		if meta.Has(ComponentRotationSpeed) {
			if !w.RotationSpeed.Set(entityID, cmd.RotationSpeed) {
				// TODO: log error
			}
		}
		if meta.Has(ComponentPlayerHitbox) {
			if !w.PlayerHitbox.Set(entityID, cmd.PlayerShape) {
				// TODO: log error
			}
		}
		if meta.Has(ComponentHealth) {
			if !w.Health.Set(entityID, cmd.Health) {
				// TODO: log error
			}
		}
		if meta.Has(ComponentCollider) {
			if !w.Collider.Set(entityID, cmd.Collider) {
				// TODO: log error
			}
		}
	}
}

func (w *World) PlayerSnapshot(id EntityID) (PlayerSnapshot, bool) {
	if !w.Entity.IsAlive(id) {
		return PlayerSnapshot{}, false
	}
	player, exist := w.playerLocation(id)
	if !exist {
		return PlayerSnapshot{}, false
	}
	return player, true
}

func (w *World) PlayerSnapshotWithView(id EntityID) (PlayerSnapshotWithView, bool) {
	if !w.Entity.IsAlive(id) {
		return PlayerSnapshotWithView{}, false
	}
	player, exist := w.playerLocation(id)
	if !exist {
		return PlayerSnapshotWithView{}, false
	}
	viewIDs, exist := w.ViewIDs.Get(id)
	if !exist {
		// TODO: log error
		return PlayerSnapshotWithView{Player: player}, true
	}

	views := make([]PlayerSnapshot, len(viewIDs))
	for i, viewID := range viewIDs {
		views[i], exist = w.playerLocation(viewID)
		if !exist {
			// TODO: log error
			continue
		}
	}
	return PlayerSnapshotWithView{Player: player, Views: views}, true
}

func (w *World) StaticEntities() []StaticEntity {
	staticEntities := make([]StaticEntity, 0)
	for entityID, collider := range w.Collider.All() {
		staticEntities = append(staticEntities, StaticEntity{
			ID:       entityID,
			Collider: collider,
		})
	}
	return staticEntities
}

func (w *World) MapInfo() MapInfo {
	return MapInfo{
		Width:  w.Width,
		Height: w.Height,
	}
}

func (w *World) playerLocation(id EntityID) (PlayerSnapshot, bool) {
	snapshot := PlayerSnapshot{ID: id}

	meta, exist := w.EntityMeta.Get(id)
	if !exist {
		return PlayerSnapshot{}, false
	}

	var position Position
	if !meta.Has(ComponentPosition) {
		position, exist = w.Position.Get(id)
		if !exist {
			// TODO: log error
		}
		snapshot.Position = position
	}

	if !meta.Has(ComponentDirection) {
		direction, exist := w.Direction.Get(id)
		if !exist {
			// TODO: log error
		}
		snapshot.Direction = direction
	}
	return snapshot, true
}

type PlayerSnapshot struct {
	ID        EntityID  `json:"id"`
	Direction Direction `json:"direction"`
	Position  Position  `json:"position"`
}

type PlayerSnapshotWithView struct {
	Player PlayerSnapshot   `json:"player"`
	Views  []PlayerSnapshot `json:"views"`
}

type StaticEntity struct {
	ID       EntityID `json:"id"`
	Collider Collider `json:"collider"`
}

type MapInfo struct {
	Width  float64
	Height float64
}
