package system

type MovementType uint8

const (
	MovementTypeAbsolute MovementType = 0
	MovementTypeRelative MovementType = 1
)

type PlayerInput struct {
	MoveVertical   float64      `json:"MoveVertical"`
	MoveHorizontal float64      `json:"MoveHorizontal"`
	LookHorizontal float64      `json:"LookHorizontal"`
	MovementType   MovementType `json:"MovementType"` // 0 = absolute, 1 = relative to player direction
	SwitchWeapon   bool         `json:"SwitchWeapon"`
	Reload         bool         `json:"Reload"`
	FastReload     bool         `json:"FastReload"`
	Fire           bool         `json:"Fire"`
	Timestamp      int64        `json:"Timestamp"`
}
