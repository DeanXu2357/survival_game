package state

import (
	"survival/internal/engine/ports"
	"survival/internal/engine/vector"
)

type Meta uint64

const (
	ComponentMeta Meta = 1 << iota
	ComponentPosition
	ComponentDirection
	ComponentMovementSpeed
	ComponentRotationSpeed
	ComponentPlayerHitbox
	ComponentHealth
	ComponentCollider
	ComponentViewIDs
	ComponentVerticalBody
	ComponentInput
	ComponentPrePosition
	ComponentProjectile
	ComponentInventory
	ComponentItemDef
	ComponentGroundItem

	PlayerMeta = ComponentMeta | ComponentPosition | ComponentDirection | ComponentMovementSpeed |
		ComponentRotationSpeed | ComponentPlayerHitbox | ComponentHealth |
		ComponentViewIDs | ComponentInput | ComponentPrePosition | ComponentInventory

	WallMeta = ComponentMeta | ComponentPosition | ComponentVerticalBody | ComponentCollider

	ProjectileMeta = ComponentMeta | ComponentPosition | ComponentDirection | ComponentProjectile

	GroundItemMeta = ComponentMeta | ComponentPosition | ComponentGroundItem
	ItemDefMeta    = ComponentMeta | ComponentItemDef
)

const (
	DefaultWallHeight        = 3.0
	DefaultWallBaseElevation = 0.0
	DefaultPlayerViewHeight  = 1.7
)

func (m Meta) Has(mask Meta) bool {
	return m&mask == mask
}

func (m Meta) Set(mask Meta) Meta {
	return m | mask
}

func (m Meta) Clear(mask Meta) Meta {
	return m &^ mask
}

type Position vector.Vector2D

type Direction float64

type PlayerHitbox struct {
	Center Position // TODO: refactor to vector2D offset design
	Radius float64
}

type MovementSpeed float64

type RotationSpeed float64

type Health int

type Collider struct {
	// Center deprecated
	Center    Position // TODO: refactor to vector2D offset design
	HalfSize  vector.Vector2D
	Direction Direction

	ShapeType ColliderShape
	Radius    float64
	Offset    vector.Vector2D
}

type ColliderShape uint8

const (
	ColliderShapeNone ColliderShape = iota
	ColliderCircle
	ColliderBox
)

func (w Collider) BoundingBox() (min vector.Vector2D, max vector.Vector2D) {
	if w.ShapeType == ColliderBox {
		return vector.Vector2D{
				X: w.Center.X - w.HalfSize.X,
				Y: w.Center.Y - w.HalfSize.Y,
			}, vector.Vector2D{
				X: w.Center.X + w.HalfSize.X,
				Y: w.Center.Y + w.HalfSize.Y,
			}
	}

	// TODO: generate circle bounding box
	//if w.ShapeType == ColliderCircle {
	//}
	return vector.Vector2D{}, vector.Vector2D{}
}

type ViewIDs []EntityID

type VerticalBody struct {
	BaseElevation float64
	Height        float64
}

type PrePosition Position

type MovementType uint8

const (
	MovementTypeAbsolute MovementType = 0
	MovementTypeRelative MovementType = 1
)

const (
	NoPickup = -1 // sentinel: no pickup action
	NoDrop   = -1 // sentinel: no drop action
)

type Input struct {
	MoveVertical   float64
	MoveHorizontal float64
	LookHorizontal float64
	MovementType   MovementType

	Fire         bool
	SwitchWeapon bool
	Reload       bool
	FastReload   bool

	PickupEntityID int64 // NoPickup (-1) = no pickup; >= 0 = target ground item entity
	DropSlotIndex  int   // NoDrop (-1) = no drop; 0-2 = weapon slot, 3+ = item slot

	Timestamp int64
}

type ProjectileData struct {
	Speed     float64 // units per second
	Range     float64 // weapon range (used to calculate TTL at spawn)
	Damage    int
	OwnerID   EntityID
	Height    float64 // fixed at 1.5 for now
	ExpiredAt uint64  // game tick at which projectile expires
}

type WeaponType uint8

const (
	WeaponTypeFist WeaponType = iota
	WeaponTypeKnife
	WeaponTypeGun
)

// WeaponSpec defines the immutable stats for a weapon type.
// Future: refactor to Flyweight pattern to share specs via pointers.
type WeaponSpec struct {
	Type     WeaponType
	Range    float64 // max range in world units
	FireRate float64 // shots per second
	Damage   int
	Speed    float64 // projectile speed in units per second
}

// WeaponSlot represents a weapon equipped in the player's weapon loadout.
type WeaponSlot struct {
	ItemDefID    EntityID // 0 = empty (slot 0 always has FistDefID)
	LastFireTick uint64   // per-weapon fire cooldown
}

func (s WeaponSlot) IsEmpty() bool { return s.ItemDefID == 0 }

// ItemSlot represents a non-weapon item in the player's inventory.
type ItemSlot struct {
	ItemDefID EntityID
	Quantity  int
}

func (s ItemSlot) IsEmpty() bool { return s.ItemDefID == 0 }

// Inventory is the unified component replacing Backpack + WeaponState.
// Weapons[0]=Fist (always), Weapons[1]=Knife slot, Weapons[2]=Gun slot.
type Inventory struct {
	Weapons            [3]WeaponSlot
	Items              [ports.ItemSlotCount]ItemSlot
	CurrentWeaponIndex int
	LastSwitchTick     uint64
	MaxItemCapacity    int
}

// ItemCount returns the number of occupied item slots.
func (inv Inventory) ItemCount() int {
	n := 0
	for _, s := range inv.Items {
		if !s.IsEmpty() {
			n++
		}
	}
	return n
}

// IsItemsFull returns true if all item slots up to MaxItemCapacity are occupied.
func (inv Inventory) IsItemsFull() bool {
	return inv.ItemCount() >= inv.MaxItemCapacity
}

// OccupiedWeaponSlots returns indices of non-empty weapon slots.
func (inv Inventory) OccupiedWeaponSlots() []int {
	var slots []int
	for i, ws := range inv.Weapons {
		if !ws.IsEmpty() {
			slots = append(slots, i)
		}
	}
	return slots
}

// FistSpec defines the default fist weapon stats.
var FistSpec = WeaponSpec{Type: WeaponTypeFist, Range: 5, FireRate: 3, Damage: 5, Speed: 50}

type ItemType uint8

const (
	ItemTypeWeapon ItemType = iota
	ItemTypeConsumable
	ItemTypeMaterial
	ItemTypeEquipment
)

type ItemDef struct {
	Name       string
	Type       ItemType
	MaxStack   int
	WeaponSpec WeaponSpec // zero-valued for non-weapons
}

type GroundItem struct {
	ItemDefID EntityID
	Quantity  int
}
