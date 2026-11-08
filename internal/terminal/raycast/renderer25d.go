package raycast

import (
	"bytes"
	"fmt"
	"math"
)

// ConeIndicator describes a flat fan-shaped ground decal centered at (X,Y)
// pointing along Direction with a frontal half-angle and max range.
type ConeIndicator struct {
	X, Y      float64
	Direction float64
	HalfAngle float64
	Range     float64
}

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
	ColorDummyNear
	ColorDummyMid
	ColorDummyFar
	ColorConeNear
	ColorConeMid
	ColorConeFar
)

var colorTo256 = map[Color]int{
	ColorBlack:     0,
	ColorFloor:     238,
	ColorCeiling:   235,
	ColorWallNear:  255,
	ColorWallMid:   245,
	ColorWallFar:   240,
	ColorDummyNear: 229, // light yellow near
	ColorDummyMid:  221, // light yellow mid
	ColorDummyFar:  178, // golden yellow far
	ColorConeNear:  196, // bright red near
	ColorConeMid:   160, // mid red
	ColorConeFar:   124, // deep red far
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

// RenderWithCones renders the scene like Render() but additionally paints
// each ConeIndicator as a red ground decal. Cones are tinted onto floor
// pixels only — wall pixels still occlude them because cone painting runs
// before wall painting via the shared logical buffer.
func (r *Renderer25D) RenderWithCones(rayHits [][]RaycastResult, playerX, playerY, playerDir float64, cones []ConeIndicator) {
	r.clearBuffer()
	if len(cones) > 0 {
		r.paintCones(rayHits, playerX, playerY, playerDir, cones)
	}
	r.paintHits(rayHits)
	r.mergeToOutput()
}

func (r *Renderer25D) Render(rayHits [][]RaycastResult) {
	r.clearBuffer()
	r.paintHits(rayHits)
	r.mergeToOutput()
}

func (r *Renderer25D) paintHits(rayHits [][]RaycastResult) {
	numRays := len(rayHits)
	if numRays == 0 {
		return
	}

	colWidth := float64(r.logicalWidth) / float64(numRays)

	for i, hits := range rayHits {
		startCol := int(float64(i) * colWidth)
		endCol := int(float64(i+1) * colWidth)
		if endCol > r.logicalWidth {
			endCol = r.logicalWidth
		}

		// Hits are sorted farthest first, so painting in-order overlays
		// nearer surfaces on top of farther ones only within their own
		// vertical span — floor/ceiling remain visible above/below.
		for _, hit := range hits {
			zDepth := hit.Distance
			if zDepth < 0.001 {
				zDepth = 0.001
			}

			yTop := r.horizon - int(((hit.BaseElevation+hit.WallHeight-r.viewHeight)/zDepth)*r.projDist)
			yBottom := r.horizon - int(((hit.BaseElevation-r.viewHeight)/zDepth)*r.projDist)

			if yTop < 0 {
				yTop = 0
			}
			if yBottom > r.logicalHeight {
				yBottom = r.logicalHeight
			}
			if yTop >= yBottom {
				continue
			}

			wallColor := r.getColor(hit.Distance, hit.ShapeType)
			for col := startCol; col < endCol; col++ {
				for y := yTop; y < yBottom; y++ {
					r.logicalBuffer[y][col] = wallColor
				}
			}
		}
	}
}

// paintCones tints floor pixels that fall inside any cone. For each ray
// (one per logical column), iterate the floor pixels (y > horizon) and
// reconstruct the world coordinate they map to, then test against every
// cone. Walls drawn afterward naturally occlude cone tint behind them.
func (r *Renderer25D) paintCones(rayHits [][]RaycastResult, playerX, playerY, playerDir float64, cones []ConeIndicator) {
	numRays := len(rayHits)
	if numRays == 0 || numRays < 2 {
		return
	}

	colWidth := float64(r.logicalWidth) / float64(numRays)
	halfFOV := FOVAngle / 2
	startAngle := playerDir - halfFOV
	angleStep := FOVAngle / float64(numRays-1)

	cosLimits := make([]float64, len(cones))
	rangeSq := make([]float64, len(cones))
	forwardX := make([]float64, len(cones))
	forwardY := make([]float64, len(cones))
	for i, c := range cones {
		cosLimits[i] = math.Cos(c.HalfAngle)
		rangeSq[i] = c.Range * c.Range
		forwardX[i] = math.Sin(c.Direction)
		forwardY[i] = -math.Cos(c.Direction)
	}

	for i := 0; i < numRays; i++ {
		rayAngle := startAngle + float64(i)*angleStep
		dx := math.Sin(rayAngle)
		dy := -math.Cos(rayAngle)
		cosDiff := math.Cos(rayAngle - playerDir) // fisheye correction
		if cosDiff < 1e-6 {
			continue
		}

		startCol := int(float64(i) * colWidth)
		endCol := int(float64(i+1) * colWidth)
		if endCol > r.logicalWidth {
			endCol = r.logicalWidth
		}

		for y := r.horizon + 1; y < r.logicalHeight; y++ {
			// Invert the wall projection (BaseElevation=0): floor pixel at
			// vertical offset (y - horizon) corresponds to projected depth
			// `r.viewHeight * projDist / (y - horizon)`. That value is the
			// view-axis distance (already fisheye-corrected); divide by
			// cosDiff to recover the actual ray distance.
			rowOffset := float64(y - r.horizon)
			viewDepth := r.viewHeight * r.projDist / rowOffset
			rayDist := viewDepth / cosDiff
			worldX := playerX + dx*rayDist
			worldY := playerY + dy*rayDist

			matched := -1
			for ci, c := range cones {
				ddx := worldX - c.X
				ddy := worldY - c.Y
				distSq := ddx*ddx + ddy*ddy
				if distSq > rangeSq[ci] {
					continue
				}
				if distSq < 1e-9 {
					matched = ci
					break
				}
				dist := math.Sqrt(distSq)
				cosAngle := (forwardX[ci]*ddx + forwardY[ci]*ddy) / dist
				if cosAngle >= cosLimits[ci] {
					matched = ci
					break
				}
			}
			if matched < 0 {
				continue
			}

			coneDist := math.Sqrt((worldX-cones[matched].X)*(worldX-cones[matched].X) +
				(worldY-cones[matched].Y)*(worldY-cones[matched].Y))
			color := coneColorFor(coneDist, cones[matched].Range)
			for col := startCol; col < endCol; col++ {
				r.logicalBuffer[y][col] = color
			}
		}
	}
}

func coneColorFor(dist, rng float64) Color {
	if rng <= 0 {
		return ColorConeFar
	}
	n := dist / rng
	if n < 0.33 {
		return ColorConeNear
	}
	if n < 0.66 {
		return ColorConeMid
	}
	return ColorConeFar
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

func (r *Renderer25D) getColor(distance float64, shapeType uint8) Color {
	normalizedDist := distance / MaxDistance

	if shapeType == shapeCircle {
		if normalizedDist < 0.33 {
			return ColorDummyNear
		} else if normalizedDist < 0.66 {
			return ColorDummyMid
		}
		return ColorDummyFar
	}

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
