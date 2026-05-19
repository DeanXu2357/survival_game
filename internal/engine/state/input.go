package state

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
