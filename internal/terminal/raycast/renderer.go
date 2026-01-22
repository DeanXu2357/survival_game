package raycast

import (
	"bytes"
	"strings"
)

var shadeChars = []rune{'█', '▓', '▒', '░', '|', ' '}

type Renderer struct {
	width  int
	height int
	buffer [][]rune
}

func NewRenderer(width, height int) *Renderer {
	buffer := make([][]rune, height)
	for i := range buffer {
		buffer[i] = make([]rune, width)
	}
	return &Renderer{
		width:  width,
		height: height,
		buffer: buffer,
	}
}

func (r *Renderer) Render(results []RaycastResult) {
	for y := range r.buffer {
		for x := range r.buffer[y] {
			r.buffer[y][x] = ' '
		}
	}

	numRays := len(results)
	if numRays == 0 {
		return
	}

	colWidth := float64(r.width) / float64(numRays)

	for i, result := range results {
		startCol := int(float64(i) * colWidth)
		endCol := int(float64(i+1) * colWidth)
		if endCol > r.width {
			endCol = r.width
		}

		wallHeight := r.calculateWallHeight(result.Distance)
		shadeChar := r.getShadeChar(result.Distance, result.Hit)

		wallTop := (r.height - wallHeight) / 2
		wallBottom := wallTop + wallHeight

		if wallTop < 0 {
			wallTop = 0
		}
		if wallBottom > r.height {
			wallBottom = r.height
		}

		for col := startCol; col < endCol; col++ {
			for row := wallTop; row < wallBottom; row++ {
				r.buffer[row][col] = shadeChar
			}
		}
	}
}

func (r *Renderer) calculateWallHeight(distance float64) int {
	if distance <= 0 {
		return r.height
	}

	normalizedDist := distance / MaxDistance
	if normalizedDist > 1 {
		normalizedDist = 1
	}

	height := int(float64(r.height) * (1 - normalizedDist*0.8))
	if height < 1 {
		height = 1
	}
	return height
}

func (r *Renderer) getShadeChar(distance float64, hit bool) rune {
	if !hit {
		return ' '
	}

	normalizedDist := distance / MaxDistance
	if normalizedDist < 0 {
		normalizedDist = 0
	}
	if normalizedDist > 1 {
		normalizedDist = 1
	}

	index := int(normalizedDist * float64(len(shadeChars)-1))
	if index >= len(shadeChars) {
		index = len(shadeChars) - 1
	}

	return shadeChars[index]
}

func (r *Renderer) WriteToBuffer(buf *bytes.Buffer) {
	for y := 0; y < r.height; y++ {
		buf.WriteString(string(r.buffer[y]))
		buf.WriteString("\033[K\r\n")
	}
}

func (r *Renderer) GetRow(row int) string {
	if row < 0 || row >= r.height {
		return strings.Repeat(" ", r.width)
	}
	return string(r.buffer[row])
}

func (r *Renderer) Width() int {
	return r.width
}

func (r *Renderer) Height() int {
	return r.height
}
