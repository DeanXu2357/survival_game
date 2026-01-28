package raycast

import (
	"bytes"
	"fmt"
)

const (
	HalfBlockUpper = '\u2580' // ▀
	FullBlockChar  = '\u2588' // █
)

type Color uint8

const (
	ColorBlack Color = iota
	ColorFloor
	ColorCeiling
	ColorWallNear
	ColorWallMid
	ColorWallFar
)

var colorTo256 = map[Color]int{
	ColorBlack:    0,
	ColorFloor:    238,
	ColorCeiling:  235,
	ColorWallNear: 255,
	ColorWallMid:  245,
	ColorWallFar:  240,
}

type ColorPair struct {
	Fg int
	Bg int
}

type Renderer25D struct {
	termWidth     int
	termHeight    int
	logicalWidth  int
	logicalHeight int
	logicalBuffer [][]Color
	outputBuffer  [][]rune
	colorBuffer   [][]ColorPair
	viewHeight    float64
	projDist      float64
	horizon       int
}

func NewRenderer25D(termWidth, termHeight int, viewHeight float64) *Renderer25D {
	logicalWidth := termWidth
	logicalHeight := termHeight * 2

	logicalBuffer := make([][]Color, logicalHeight)
	for i := range logicalBuffer {
		logicalBuffer[i] = make([]Color, logicalWidth)
	}

	outputBuffer := make([][]rune, termHeight)
	colorBuffer := make([][]ColorPair, termHeight)
	for i := range outputBuffer {
		outputBuffer[i] = make([]rune, termWidth)
		colorBuffer[i] = make([]ColorPair, termWidth)
	}

	return &Renderer25D{
		termWidth:     termWidth,
		termHeight:    termHeight,
		logicalWidth:  logicalWidth,
		logicalHeight: logicalHeight,
		logicalBuffer: logicalBuffer,
		outputBuffer:  outputBuffer,
		colorBuffer:   colorBuffer,
		viewHeight:    viewHeight,
		horizon:       logicalHeight / 2,

		// FOV of 90 degrees, projDist = width / (2 * tan(FOV/2)) = width / 2
		// logicalWidth ~= 53 degrees, logicalWidth/4 ~= 126 degrees
		projDist: float64(logicalWidth) / 2,
	}
}

func (r *Renderer25D) Render(results []RaycastResult) {
	r.clearBuffer()

	numRays := len(results)
	if numRays == 0 {
		r.mergeToOutput()
		return
	}

	colWidth := float64(r.logicalWidth) / float64(numRays)

	for i, result := range results {
		startCol := int(float64(i) * colWidth)
		endCol := int(float64(i+1) * colWidth)
		if endCol > r.logicalWidth {
			endCol = r.logicalWidth
		}

		if !result.Hit {
			for col := startCol; col < endCol; col++ {
				r.drawColumn(col, r.horizon, r.horizon)
			}
			continue
		}

		wallHeight := result.WallHeight
		baseElev := result.BaseElevation
		zDepth := result.Distance

		if zDepth < 0.001 {
			zDepth = 0.001
		}

		yTop := r.horizon - int(((baseElev+wallHeight-r.viewHeight)/zDepth)*r.projDist)
		yBottom := r.horizon - int(((baseElev-r.viewHeight)/zDepth)*r.projDist)

		wallColor := r.getWallColor(result.Distance)

		for col := startCol; col < endCol; col++ {
			r.drawColumnWithWall(col, yTop, yBottom, wallColor)
		}
	}

	r.mergeToOutput()
}

func (r *Renderer25D) clearBuffer() {
	for y := 0; y < r.logicalHeight; y++ {
		for x := 0; x < r.logicalWidth; x++ {
			if y < r.horizon {
				r.logicalBuffer[y][x] = ColorCeiling
			} else {
				r.logicalBuffer[y][x] = ColorFloor
			}
		}
	}
}

func (r *Renderer25D) drawColumn(col, yTop, yBottom int) {
	for y := 0; y < r.logicalHeight; y++ {
		if y < r.horizon {
			r.logicalBuffer[y][col] = ColorCeiling
		} else {
			r.logicalBuffer[y][col] = ColorFloor
		}
	}
}

func (r *Renderer25D) drawColumnWithWall(col, yTop, yBottom int, wallColor Color) {
	if yTop < 0 {
		yTop = 0
	}
	if yBottom > r.logicalHeight {
		yBottom = r.logicalHeight
	}

	for y := 0; y < r.logicalHeight; y++ {
		if y < yTop {
			r.logicalBuffer[y][col] = ColorCeiling
		} else if y >= yTop && y < yBottom {
			r.logicalBuffer[y][col] = wallColor
		} else {
			r.logicalBuffer[y][col] = ColorFloor
		}
	}
}

func (r *Renderer25D) getWallColor(distance float64) Color {
	normalizedDist := distance / MaxDistance
	if normalizedDist < 0.33 {
		return ColorWallNear
	} else if normalizedDist < 0.66 {
		return ColorWallMid
	}
	return ColorWallFar
}

func (r *Renderer25D) WriteToBuffer(buf *bytes.Buffer) {
	for termRow := 0; termRow < r.termHeight; termRow++ {
		upperRow := termRow * 2
		lowerRow := termRow*2 + 1

		for col := 0; col < r.termWidth; col++ {
			upperColor := r.logicalBuffer[upperRow][col]
			lowerColor := r.logicalBuffer[lowerRow][col]

			fg := colorTo256[upperColor]
			bg := colorTo256[lowerColor]

			// TODO: Investigate scanline artifacts (horizontal gaps) appearing between rows.
			// These visual glitches are likely caused by terminal line height or font rendering settings.
			// A potential fix involves adjusting background color logic or enforcing line spacing.
			buf.WriteString(fmt.Sprintf("\033[38;5;%dm\033[48;5;%dm%c", fg, bg, HalfBlockUpper))
		}
		buf.WriteString("\033[0m\033[K\r\n")
	}
}

func (r *Renderer25D) Width() int {
	return r.termWidth
}

func (r *Renderer25D) Height() int {
	return r.termHeight
}

func (r *Renderer25D) SetViewHeight(height float64) {
	r.viewHeight = height
}

func (r *Renderer25D) mergeToOutput() {
	for termRow := 0; termRow < r.termHeight; termRow++ {
		upperRow := termRow * 2
		lowerRow := termRow*2 + 1

		for col := 0; col < r.termWidth; col++ {
			upperColor := r.logicalBuffer[upperRow][col]
			lowerColor := r.logicalBuffer[lowerRow][col]

			r.outputBuffer[termRow][col] = HalfBlockUpper
			r.colorBuffer[termRow][col] = ColorPair{
				Fg: colorTo256[upperColor],
				Bg: colorTo256[lowerColor],
			}
		}
	}
}

func (r *Renderer25D) GetOutputBuffer() [][]rune {
	return r.outputBuffer
}

func (r *Renderer25D) GetColorBuffer() [][]ColorPair {
	return r.colorBuffer
}

func (r *Renderer25D) WriteWithOverlay(buf *bytes.Buffer) {
	for termRow := 0; termRow < r.termHeight; termRow++ {
		for col := 0; col < r.termWidth; col++ {
			ch := r.outputBuffer[termRow][col]
			cp := r.colorBuffer[termRow][col]
			buf.WriteString(fmt.Sprintf("\033[38;5;%dm\033[48;5;%dm%c", cp.Fg, cp.Bg, ch))
		}
		buf.WriteString("\033[0m\033[K\r\n")
	}
}
