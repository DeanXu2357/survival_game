package weapons

import (
	"survival/internal/engine/vector"
)

type Projectile struct {
	ID        string
	Position  vector.Vector2D
	Direction vector.Vector2D
	Speed     float64
	Range     float64
	Damage    int
	OwnerID   string
}

func (p *Projectile) GetPosition() vector.Vector2D {
	return p.Position
}
