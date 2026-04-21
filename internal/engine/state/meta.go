package state

type Meta uint64

const (
	ComponentMeta Meta = 1 << iota
	ComponentPosition
	ComponentDirection
	ComponentMovementSpeed
	ComponentRotationSpeed
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
		ComponentRotationSpeed | ComponentCollider | ComponentHealth |
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
