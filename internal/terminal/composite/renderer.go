package composite

import (
	"bytes"
	"strings"

	"survival/internal/engine/ports"
	"survival/internal/terminal/debug"
	"survival/internal/terminal/raycast"
)

type Renderer struct {
	width     int
	height    int
	buffer    [][]rune
	rowBuffer []byte

	raycast  *raycast.Renderer
	debugMap *debug.MapRenderer

	mainWidth  int
	debugWidth int
	sepWidth   int
}

func NewRenderer(width, height, mainWidth, debugWidth, sepWidth int, viewRadius float64) *Renderer {
	buffer := make([][]rune, height)
	for i := range buffer {
		buffer[i] = make([]rune, width)
	}

	debugHeight := height

	return &Renderer{
		width:      width,
		height:     height,
		buffer:     buffer,
		rowBuffer:  make([]byte, 0, width*4),
		raycast:    raycast.NewRenderer(mainWidth, height),
		debugMap:   debug.NewMapRenderer(debugWidth, debugHeight, viewRadius),
		mainWidth:  mainWidth,
		debugWidth: debugWidth,
		sepWidth:   sepWidth,
	}
}

func (r *Renderer) Render(results []raycast.RaycastResult, playerX, playerY, playerDir float64, colliders []ports.Collider) {
	r.raycast.Render(results)
	//r.debugMap.Render(playerX, playerY, playerDir, colliders)
	r.compose()
}

func (r *Renderer) compose() {
	separator := strings.Repeat(" ", r.sepWidth)

	for row := 0; row < r.height; row++ {
		mainRow := r.raycast.GetRow(row)
		//debugRow := r.debugMap.GetRow(row)

		col := 0
		for _, ch := range mainRow {
			if col < r.width {
				r.buffer[row][col] = ch
				col++
			}
		}

		for _, ch := range separator {
			if col < r.width {
				r.buffer[row][col] = ch
				col++
			}
		}

		//for _, ch := range debugRow {
		//	if col < r.width {
		//		r.buffer[row][col] = ch
		//		col++
		//	}
		//}

		for col < r.width {
			r.buffer[row][col] = ' '
			col++
		}
	}
}

func (r *Renderer) WriteToBuffer(buf *bytes.Buffer) {
	for row := 0; row < r.height; row++ {
		buf.WriteString(string(r.buffer[row]))
		buf.WriteString("\033[K\r\n")
	}
}

func (r *Renderer) GetRow(row int) string {
	if row < 0 || row >= r.height {
		return strings.Repeat(" ", r.width)
	}
	return string(r.buffer[row])
}
