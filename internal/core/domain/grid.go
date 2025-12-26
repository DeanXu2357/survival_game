package domain

import (
	"iter"
	"math"
)

type Grid struct {
	cellSize      float64
	width, height int
	cellSlice     []GridCell
}

func NewGrid(cellSize float64, width, height int) *Grid {
	return &Grid{
		cellSize:  cellSize,
		width:     width,
		height:    height,
		cellSlice: make([]GridCell, width*height),
	}
}
type GridCell struct {
	entries []GridEntry
}

type GridEntry struct {
	EntityID EntityID
	Layer    LayerMask
}

type Bounds struct {
	MinX, MinY float64
	MaxX, MaxY float64
}

// Add adds an entity with given bounds and layer to the grid.
// It returns the list of grid cell indexes the entity was added to.
func (g *Grid) Add(id EntityID, bounds Bounds, layer LayerMask) []int {
	minGX, minGY := g.GridCoord(bounds.MinX, bounds.MinY)
	maxGX, maxGY := g.GridCoord(bounds.MaxX, bounds.MaxY)

	indexes := make([]int, 0, (maxGX-minGX+1)*(maxGY-minGY+1))

	entry := GridEntry{EntityID: id, Layer: layer}

	for gx := minGX; gx <= maxGX; gx++ {
		for gy := minGY; gy <= maxGY; gy++ {
			index := g.GridIndex(gx, gy)
			if index != -1 {
				g.cellSlice[index].entries = append(g.cellSlice[index].entries, entry)
				indexes = append(indexes, index)
			}
		}
	}
	return indexes
}

func (g *Grid) Remove(indexes []uint64, id EntityID) {
	for _, cellIDX := range indexes {
		cell := &g.cellSlice[cellIDX]

		for i, entry := range cell.entries {
			if entry.EntityID == id {
				last := len(cell.entries) - 1
				cell.entries[i] = cell.entries[last]
				cell.entries = cell.entries[:last]
				break
			}
		}
	}
}

func (g *Grid) AllCells() iter.Seq2[int, *GridCell] {
	return func(yield func(int, *GridCell) bool) {
		for i := range g.cellSlice {
			if !yield(i, &g.cellSlice[i]) {
				return
			}
		}
	}
}

// CellsInBounds yields all grid cells within the given world-coord bounds.
func (g *Grid) CellsInBounds(bounds Bounds) iter.Seq2[int, *GridCell] {
	minGX, minGY := g.GridCoord(bounds.MinX, bounds.MinY)
	maxGX, maxGY := g.GridCoord(bounds.MaxX, bounds.MaxY)

	return func(yield func(int, *GridCell) bool) {
		for gx := minGX; gx <= maxGX; gx++ {
			for gy := minGY; gy <= maxGY; gy++ {
				index := g.GridIndex(gx, gy)
				if index != -1 {
					if !yield(index, &g.cellSlice[index]) {
						return
					}
				}
			}
		}
	}
}

// GridCoord converts world coordinates to grid coordinates.
// float(-0.5) floored is -1, so this works correctly for negative coordinates as well.
func (g *Grid) GridCoord(x, y float64) (int, int) {
	gx := int(math.Floor(x / g.cellSize))
	gy := int(math.Floor(y / g.cellSize))
	return gx, gy
}

// GridIndex converts grid coordinates to a linear index.
// Returns -1 if out of bounds.
func (g *Grid) GridIndex(gx, gy int) int {
	if gx < 0 || gy < 0 || gx >= g.width || gy >= g.height {
		return -1
	}
	return gy*g.width + gx
}

// GridCoordFromIndex converts a linear index to grid coordinates.
// Returns -1, -1 if out of bounds.
func (g *Grid) GridCoordFromIndex(index int) (int, int) {
	if index < 0 || index >= len(g.cellSlice) {
		return -1, -1
	}

	y := index / g.width
	x := index % g.width
	return x, y
}

func (g *Grid) GridIndexFromWorld(x, y float64) int {
	gx, gy := g.GridCoord(x, y)
	return g.GridIndex(gx, gy)
}

func (g *Grid) WorldCoordOfCellLeftTop(gx, gy int) (float64, float64) {
	x := float64(gx) * g.cellSize
	y := float64(gy) * g.cellSize
	return x, y
}

func (g *Grid) WorldCoordOfCellCenter(gx, gy int) (float64, float64) {
	x := (float64(gx) + 0.5) * g.cellSize
	y := (float64(gy) + 0.5) * g.cellSize
	return x, y
}

func FilterByLayer(cells iter.Seq[GridEntry], layer LayerMask) iter.Seq[GridEntry] {
	return func(yield func(GridEntry) bool) {
		for cell := range cells {
			if cell.Layer.Has(layer) {
				if !yield(cell) {
					return
				}
			}
		}
	}
}
