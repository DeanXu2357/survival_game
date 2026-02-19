package state

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
