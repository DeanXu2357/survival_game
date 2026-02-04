package state

import (
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

	PlayerMeta = ComponentMeta | ComponentPosition | ComponentDirection | ComponentMovementSpeed |
		ComponentRotationSpeed | ComponentPlayerHitbox | ComponentHealth |
		ComponentViewIDs | ComponentInput | ComponentPrePosition

	WallMeta = ComponentMeta | ComponentPosition | ComponentVerticalBody | ComponentCollider
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

type Input struct {
	MoveVertical   float64
	MoveHorizontal float64
	LookHorizontal float64
	MovementType   MovementType

	Fire         bool
	SwitchWeapon bool
	Reload       bool
	FastReload   bool

	Timestamp int64
}
