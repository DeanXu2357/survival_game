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
	ID       EntityID `json:"id"`
	Collider Collider `json:"collider"`
}

type GroundItemSnapshot struct {
	ID         EntityID
	Position   Position
	GroundItem GroundItem
}

type MapInfo struct {
	Width  float64
	Height float64
}
