package game

type Vector2D struct {
	X float64
	Y float64
}

type Player struct {
	ID            string
	Position      *Vector2D
	Rotation      float64
	Health        int
	IsAlive       bool
	Inventory     *Inventory
	CurrentWeapon Weapon
}

type Wall struct {
	ID       string
	Position Vector2D
	Width    float64
	Height   float64
}

func (w *Wall) GetPosition() Vector2D {
	return w.Position
}

type Projectile struct {
	ID        string
	Position  Vector2D
	Direction Vector2D
	Speed     float64
	Range     float64
	Damage    int
	OwnerID   string
}

func (p *Projectile) GetPosition() Vector2D {
	return p.Position
}

type State struct {
	Players     map[string]*Player
	Walls       []*Wall
	Projectiles []*Projectile
}

func NewGameState() *State {
	return &State{
		Players:     make(map[string]*Player),
		Walls:       make([]*Wall, 0),
		Projectiles: make([]*Projectile, 0),
	}
}

type MapObject interface {
	GetPosition() Vector2D
}

type GridCoord struct {
	X int
	Y int
}

type Grid struct {
	CellSize float64
	Cells    map[GridCoord][]*MapObject
}

func (g *Grid) getCoord(worldPos Vector2D) GridCoord {
	return GridCoord{
		X: int(worldPos.X / g.CellSize),
		Y: int(worldPos.Y / g.CellSize),
	}
}
