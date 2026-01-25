package debug

import (
	"math"

	"survival/internal/engine/ports"
)

const (
	charEmpty      = ' '
	charWall       = '█'
	charBorderH    = '─'
	charBorderV    = '│'
	charBorderTL   = '┌'
	charBorderTR   = '┐'
	charBorderBL   = '└'
	charBorderBR   = '┘'
	charPlayerUp   = '^'
	charPlayerDown = 'v'
	charPlayerLeft = '<'
	charPlayerRight = '>'
)

type MapRenderer struct {
	width      int
	height     int
	buffer     [][]rune
	viewRadius float64
}

func NewMapRenderer(width, height int, viewRadius float64) *MapRenderer {
	buffer := make([][]rune, height)
	for i := range buffer {
		buffer[i] = make([]rune, width)
	}
	return &MapRenderer{
		width:      width,
		height:     height,
		buffer:     buffer,
		viewRadius: viewRadius,
	}
}

func (m *MapRenderer) Render(playerX, playerY, playerDir float64, colliders []ports.Collider) {
	m.clear()
	m.drawBorder()
	m.renderWalls(colliders, playerX, playerY)
	m.renderPlayer(playerDir)
}

func (m *MapRenderer) clear() {
	for y := range m.buffer {
		for x := range m.buffer[y] {
			m.buffer[y][x] = charEmpty
		}
	}
}

func (m *MapRenderer) drawBorder() {
	m.buffer[0][0] = charBorderTL
	m.buffer[0][m.width-1] = charBorderTR
	m.buffer[m.height-1][0] = charBorderBL
	m.buffer[m.height-1][m.width-1] = charBorderBR

	for x := 1; x < m.width-1; x++ {
		m.buffer[0][x] = charBorderH
		m.buffer[m.height-1][x] = charBorderH
	}

	for y := 1; y < m.height-1; y++ {
		m.buffer[y][0] = charBorderV
		m.buffer[y][m.width-1] = charBorderV
	}
}

func (m *MapRenderer) worldToScreen(worldX, worldY, centerX, centerY float64) (screenX, screenY int, inBounds bool) {
	innerWidth := float64(m.width - 2)
	innerHeight := float64(m.height - 2)
	viewSize := m.viewRadius * 2

	relX := worldX - (centerX - m.viewRadius)
	relY := worldY - (centerY - m.viewRadius)

	scaleX := innerWidth / viewSize
	scaleY := innerHeight / viewSize

	screenX = int(relX*scaleX) + 1
	screenY = int(relY*scaleY) + 1

	inBounds = screenX >= 1 && screenX < m.width-1 && screenY >= 1 && screenY < m.height-1
	return
}

func (m *MapRenderer) renderWalls(colliders []ports.Collider, centerX, centerY float64) {
	for _, collider := range colliders {
		m.renderCollider(collider, centerX, centerY)
	}
}

func (m *MapRenderer) renderCollider(collider ports.Collider, centerX, centerY float64) {
	minX := collider.X - collider.HalfX
	maxX := collider.X + collider.HalfX
	minY := collider.Y - collider.HalfY
	maxY := collider.Y + collider.HalfY

	viewMinX := centerX - m.viewRadius
	viewMaxX := centerX + m.viewRadius
	viewMinY := centerY - m.viewRadius
	viewMaxY := centerY + m.viewRadius

	if maxX < viewMinX || minX > viewMaxX || maxY < viewMinY || minY > viewMaxY {
		return
	}

	innerWidth := float64(m.width - 2)
	innerHeight := float64(m.height - 2)
	viewSize := m.viewRadius * 2

	scaleX := innerWidth / viewSize
	scaleY := innerHeight / viewSize

	screenMinX := int((minX-viewMinX)*scaleX) + 1
	screenMaxX := int((maxX-viewMinX)*scaleX) + 1
	screenMinY := int((minY-viewMinY)*scaleY) + 1
	screenMaxY := int((maxY-viewMinY)*scaleY) + 1

	if screenMinX < 1 {
		screenMinX = 1
	}
	if screenMaxX >= m.width-1 {
		screenMaxX = m.width - 2
	}
	if screenMinY < 1 {
		screenMinY = 1
	}
	if screenMaxY >= m.height-1 {
		screenMaxY = m.height - 2
	}

	for y := screenMinY; y <= screenMaxY; y++ {
		for x := screenMinX; x <= screenMaxX; x++ {
			m.buffer[y][x] = charWall
		}
	}
}

func (m *MapRenderer) renderPlayer(playerDir float64) {
	centerX := m.width / 2
	centerY := m.height / 2

	playerChar := m.getDirectionChar(playerDir)
	m.buffer[centerY][centerX] = playerChar
}

func (m *MapRenderer) getDirectionChar(dir float64) rune {
	dir = math.Mod(dir, 2*math.Pi)
	if dir < 0 {
		dir += 2 * math.Pi
	}

	eighth := math.Pi / 4

	// Backend convention: 0=up, π/2=right, π=down, 3π/2=left
	if dir < eighth || dir >= 7*eighth {
		return charPlayerUp
	} else if dir < 3*eighth {
		return charPlayerRight
	} else if dir < 5*eighth {
		return charPlayerDown
	} else {
		return charPlayerLeft
	}
}

func (m *MapRenderer) GetRow(row int) string {
	if row < 0 || row >= m.height {
		return ""
	}
	return string(m.buffer[row])
}

func (m *MapRenderer) Width() int {
	return m.width
}

func (m *MapRenderer) Height() int {
	return m.height
}
