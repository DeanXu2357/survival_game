package game

import "survival/internal/vector"

type Grid struct {
	CellSize float64
	Cells    map[GridCoord][]MapObject
}

func (g *Grid) getCoord(worldPos vector.Vector2D) GridCoord {
	return GridCoord{
		X: int(worldPos.X / g.CellSize),
		Y: int(worldPos.Y / g.CellSize),
	}
}

func (g *Grid) AddObject(obj MapObject) {
	coord := g.getCoord(obj.Position())
	if _, exists := g.Cells[coord]; !exists {
		g.Cells[coord] = make([]MapObject, 0)
	}
	g.Cells[coord] = append(g.Cells[coord], obj)
}

func (g *Grid) RemoveObject(obj MapObject) {
	coord := g.getCoord(obj.Position())
	if objects, exists := g.Cells[coord]; exists {
		for i, o := range objects {
			if o.ID() == obj.ID() {
				g.Cells[coord] = append(objects[:i], objects[i+1:]...)
				break
			}
		}
		if len(g.Cells[coord]) == 0 {
			delete(g.Cells, coord)
		}
	}
}

func (g *Grid) NearbyPositions(worldPos ...vector.Vector2D) []MapObject {
	nearby := make([]MapObject, 0)

	// FIXME: This function should handle the positions crossing the grid boundaries
	for _, pos := range worldPos {
		coord := g.getCoord(pos)

		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				neighborCoord := GridCoord{X: coord.X + dx, Y: coord.Y + dy}
				if objects, exists := g.Cells[neighborCoord]; exists {
					nearby = append(nearby, objects...)
				}
			}
		}
	}

	return nearby
}
