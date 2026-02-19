package state

import (
	"fmt"
	"log"
	"sync"

	"survival/internal/engine/ports"
)

type World struct {
	Entity *EntityManager

	EntityMeta    ComponentManager[Meta]
	Position      ComponentManager[Position]
	PrePosition   ComponentManager[PrePosition]
	Direction     ComponentManager[Direction]
	MovementSpeed ComponentManager[MovementSpeed]
	RotationSpeed ComponentManager[RotationSpeed]

	ViewIDs      ComponentManager[ViewIDs]
	PlayerHitbox ComponentManager[PlayerHitbox]

	Health   ComponentManager[Health]
	Collider ComponentManager[Collider]

	VerticalBody ComponentManager[VerticalBody]

	Projectile ComponentManager[ProjectileData]
	Inventory  ComponentManager[Inventory]
	ItemDef    ComponentManager[ItemDef]
	GroundItem ComponentManager[GroundItem]

	Input          ComponentManager[Input]
	inputMapBuffer map[EntityID]Input
	inputMutex     *sync.Mutex

	Grid Grid

	buf *CommandBuffer

	Width, Height float64
}

func NewWorld(gridCellSize float64, gridWidth, gridHeight int) *World {
	return &World{
		Entity:         NewEntityManager(),
		EntityMeta:     *NewComponentManager[Meta](), // TODO: refactor this, use pointer or not
		Position:       *NewComponentManager[Position](),
		PrePosition:    *NewComponentManager[PrePosition](),
		Direction:      *NewComponentManager[Direction](),
		MovementSpeed:  *NewComponentManager[MovementSpeed](),
		RotationSpeed:  *NewComponentManager[RotationSpeed](),
		ViewIDs:        *NewComponentManager[ViewIDs](),
		PlayerHitbox:   *NewComponentManager[PlayerHitbox](),
		Health:         *NewComponentManager[Health](),
		Collider:       *NewComponentManager[Collider](),
		VerticalBody:   *NewComponentManager[VerticalBody](),
		Projectile:     *NewComponentManager[ProjectileData](),
		Inventory:      *NewComponentManager[Inventory](),
		ItemDef:        *NewComponentManager[ItemDef](),
		GroundItem:     *NewComponentManager[GroundItem](),
		Input:          *NewComponentManager[Input](),
		inputMapBuffer: make(map[EntityID]Input),
		inputMutex:     &sync.Mutex{},
		Grid:           *NewGrid(gridCellSize, gridWidth, gridHeight),
		buf:            NewCommandBuffer(),
		Width:          0,
		Height:         0,
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

	fmt.Printf("Allocated Player EntityID %d\n", id)

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
			Inventory:     DefaultInventory(cfg.FistDefID, ports.ItemSlotCount),
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
	FistDefID     EntityID
}

func (w *World) UpdatePlayer(id EntityID, player UpdatePlayer) {
	log.Printf("Queue UpdatePlayer command for EntityID %d", id)
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
		PrePosition:   player.PrePosition,
		Inventory:     player.Inventory,
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
	PrePosition
	Inventory
}

type CreateProjectile struct {
	Position  Position
	Direction Direction
	ProjectileData
}

// CreateProjectileEntity allocates a projectile entity and queues its components via command buffer.
func (w *World) CreateProjectileEntity(cfg CreateProjectile) (EntityID, bool) {
	id, ok := w.Entity.Alloc()
	if !ok {
		return 0, false
	}

	w.buf.Push(WorldCommand{
		EntityID:       id,
		UpdateMeta:     ProjectileMeta,
		Position:       cfg.Position,
		Direction:      cfg.Direction,
		Meta:           ProjectileMeta,
		ProjectileData: cfg.ProjectileData,
	})

	return id, true
}

// QueueDestroyEntity queues an entity for destruction in the next ApplyCommands call.
func (w *World) QueueDestroyEntity(id EntityID) {
	w.buf.Push(WorldCommand{
		EntityID: id,
		Destroy:  true,
	})
}

func (w *World) ApplyCommands() {
	for !w.buf.IsEmpty() {
		cmd, ok := w.buf.Pop()
		if !ok {
			log.Printf("[Warning] ApplyCommands: CommandBuffer is empty unexpectedly")
			continue
		}

		entityID := cmd.EntityID
		if !w.Entity.IsAlive(entityID) {
			log.Printf("ApplyCommands: EntityID %d is not alive, skipping command", entityID)
			continue
		}

		if cmd.Destroy {
			w.destroyEntity(entityID)
			continue
		}

		log.Printf("Applying command for EntityID %d with UpdateMeta %b", entityID, cmd.UpdateMeta)

		if cmd.UpdateMeta.Has(ComponentMeta) {
			if !w.EntityMeta.Upsert(entityID, cmd.Meta) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentPosition) {
			if !w.Position.Upsert(entityID, cmd.Position) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentDirection) {
			if !w.Direction.Upsert(entityID, cmd.Direction) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentMovementSpeed) {
			if !w.MovementSpeed.Upsert(entityID, cmd.MovementSpeed) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentRotationSpeed) {
			if !w.RotationSpeed.Upsert(entityID, cmd.RotationSpeed) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentPlayerHitbox) {
			if !w.PlayerHitbox.Upsert(entityID, cmd.PlayerShape) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentHealth) {
			if !w.Health.Upsert(entityID, cmd.Health) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentCollider) {
			if !w.Collider.Upsert(entityID, cmd.Collider) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentVerticalBody) {
			if !w.VerticalBody.Upsert(entityID, cmd.VerticalBody) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentInput) {
			if !w.Input.Upsert(entityID, cmd.Input) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentPrePosition) {
			if !w.PrePosition.Upsert(entityID, cmd.PrePosition) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentProjectile) {
			if !w.Projectile.Upsert(entityID, cmd.ProjectileData) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentInventory) {
			if !w.Inventory.Upsert(entityID, cmd.Inventory) {
				// TODO: log error
			}
		}
		if cmd.UpdateMeta.Has(ComponentGroundItem) {
			if !w.GroundItem.Upsert(entityID, cmd.GroundItem) {
				// TODO: log error
			}
		}
	}
}

// destroyEntity removes all components for an entity and frees its EntityID.
func (w *World) destroyEntity(id EntityID) {
	// TODO: optimize this by keeping track of which components an entity has in its Meta
	w.EntityMeta.Remove(id)
	w.Position.Remove(id)
	w.PrePosition.Remove(id)
	w.Direction.Remove(id)
	w.MovementSpeed.Remove(id)
	w.RotationSpeed.Remove(id)
	w.ViewIDs.Remove(id)
	w.PlayerHitbox.Remove(id)
	w.Health.Remove(id)
	w.Collider.Remove(id)
	w.VerticalBody.Remove(id)
	w.Projectile.Remove(id)
	w.Inventory.Remove(id)
	w.ItemDef.Remove(id)
	w.GroundItem.Remove(id)
	w.Input.Remove(id)
	w.Entity.Free(id)
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
		log.Printf("PlayerSnapshotWithView: failed to get player location for EntityID %d", id)
		return PlayerSnapshotWithView{}, false
	}
	viewIDs, exist := w.ViewIDs.Get(id)
	if !exist {
		// TODO: log error
		log.Printf("PlayerSnapshotWithView: no ViewIDs component for EntityID %d", id)
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
		entity := StaticEntity{
			ID:       entityID,
			Collider: collider,
		}
		if vertBody, ok := w.VerticalBody.Get(entityID); ok {
			entity.VerticalBody = vertBody
			entity.HasVerticalBody = true
		}
		staticEntities = append(staticEntities, entity)
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
		log.Printf("playerLocation: no Meta component for EntityID %d", id)
		return PlayerSnapshot{}, false
	}

	if meta.Has(ComponentPosition) {
		if position, exist := w.Position.Get(id); exist {
			snapshot.Position = position
		} else {
			log.Printf("playerLocation: no Position component for EntityID %d", id)
			// TODO: log error
		}
	}

	if meta.Has(ComponentDirection) {
		if direction, exist := w.Direction.Get(id); exist {
			snapshot.Direction = direction
		} else {
			log.Printf("playerLocation: no Direction component for EntityID %d", id)
			// TODO: log error
		}
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
	ID              EntityID     `json:"id"`
	Collider        Collider     `json:"collider"`
	VerticalBody    VerticalBody `json:"vertical_body"`
	HasVerticalBody bool         `json:"has_vertical_body"`
}

type MapInfo struct {
	Width  float64
	Height float64
}

// DefaultInventory returns the initial inventory for a new player.
// fistDefID is the EntityID of the Fist item definition (always occupies slot 0).
func DefaultInventory(fistDefID EntityID, maxItemCapacity int) Inventory {
	return Inventory{
		Weapons: [3]WeaponSlot{
			{ItemDefID: fistDefID}, // Slot 0: Fist (always occupied)
			{},                     // Slot 1: Knife (empty)
			{},                     // Slot 2: Gun (empty)
		},
		CurrentWeaponIndex: 0, // default to Fist
		MaxItemCapacity:    maxItemCapacity,
	}
}

type CreateGroundItem struct {
	Position  Position
	ItemDefID EntityID
	Quantity  int
}

// CreateGroundItemEntity allocates a ground item entity referencing an item definition config entity.
func (w *World) CreateGroundItemEntity(cfg CreateGroundItem) (EntityID, bool) {
	id, ok := w.Entity.Alloc()
	if !ok {
		return 0, false
	}

	w.buf.Push(WorldCommand{
		EntityID:   id,
		UpdateMeta: GroundItemMeta,
		Position:   cfg.Position,
		Meta:       GroundItemMeta,
		GroundItem: GroundItem{ItemDefID: cfg.ItemDefID, Quantity: cfg.Quantity},
	})

	return id, true
}

// CreateItemDefEntity allocates an item definition config entity (persists forever).
// ItemDefs are immutable config data, so they are written directly (no command buffer).
func (w *World) CreateItemDefEntity(def ItemDef) (EntityID, bool) {
	id, ok := w.Entity.Alloc()
	if !ok {
		return 0, false
	}

	w.EntityMeta.Upsert(id, ItemDefMeta)
	w.ItemDef.Upsert(id, def)

	return id, true
}

// SetInput buffers the input for an entity.
// Inputs are merged if multiple inputs are set before SyncInputBuffer is called.
// This method is thread-safe.
func (w *World) SetInput(entityID EntityID, input Input) {
	w.inputMutex.Lock()
	defer w.inputMutex.Unlock()

	old := w.inputMapBuffer[entityID]

	dropSlotIndex := old.DropSlotIndex
	if input.DropSlotIndex != NoDrop {
		dropSlotIndex = input.DropSlotIndex
	}

	pickupEntityID := old.PickupEntityID
	if input.PickupEntityID != NoPickup {
		pickupEntityID = input.PickupEntityID
	}

	w.inputMapBuffer[entityID] = Input{
		MoveVertical:   input.MoveVertical,
		MoveHorizontal: input.MoveHorizontal,
		LookHorizontal: input.LookHorizontal,
		MovementType:   input.MovementType,
		Fire:           old.Fire || input.Fire,
		SwitchWeapon:   old.SwitchWeapon || input.SwitchWeapon,
		Reload:         old.Reload || input.Reload,
		FastReload:     old.FastReload || input.FastReload,
		PickupEntityID: pickupEntityID,
		DropSlotIndex:  dropSlotIndex,
		Timestamp:      input.Timestamp,
	}
}

// SyncInputBuffer flushes the input buffer into the main Input component manager.
// Should be called at the start of each simulation tick.
func (w *World) SyncInputBuffer() {
	w.inputMutex.Lock()
	defer w.inputMutex.Unlock()

	for entityID, input := range w.inputMapBuffer {
		w.Input.Upsert(entityID, input)

		// Clear the buffer after syncing
		delete(w.inputMapBuffer, entityID)
	}
}
