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

type IndicatorShape uint8

const (
	IndicatorShapeNone IndicatorShape = iota
	IndicatorShapeCone
)

// ThreatIndicator is a renderable cue (e.g. a turret's attack cone)
// emitted by an AI behavior. Shape selects which geometry fields are
// meaningful — for ShapeCone, the renderer reads Direction, HalfAngle,
// and Range.
type ThreatIndicator struct {
	EntityID  EntityID
	Shape     IndicatorShape
	X, Y      float64
	Direction float64
	HalfAngle float64
	Range     float64
}

type MapInfo struct {
	Width  float64
	Height float64
}
