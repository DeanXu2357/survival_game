package ui

import (
	"bytes"
	"fmt"

	"survival/internal/terminal/raycast"
)

type UILayer struct {
	width            int
	height           int
	crosshairEnabled bool
	weaponEnabled    bool
	hudEnabled       bool
	health           int
	ammo             int
	maxAmmo          int
}

func NewUILayer(width, height int) *UILayer {
	return &UILayer{
		width:            width,
		height:           height,
		crosshairEnabled: true,
		weaponEnabled:    true,
		hudEnabled:       true,
		health:           100,
		ammo:             12,
		maxAmmo:          12,
	}
}

func (u *UILayer) SetCrosshairEnabled(enabled bool) {
	u.crosshairEnabled = enabled
}

func (u *UILayer) SetWeaponEnabled(enabled bool) {
	u.weaponEnabled = enabled
}

func (u *UILayer) SetHUDEnabled(enabled bool) {
	u.hudEnabled = enabled
}

func (u *UILayer) SetHealth(health int) {
	u.health = health
}

func (u *UILayer) SetAmmo(ammo, maxAmmo int) {
	u.ammo = ammo
	u.maxAmmo = maxAmmo
}

func (u *UILayer) Overlay(buffer [][]rune, colors [][]raycast.ColorPair) {
	if u.crosshairEnabled {
		u.drawCrosshair(buffer, colors)
	}

	if u.hudEnabled {
		u.drawHUD(buffer, colors)
	}

	if u.weaponEnabled {
		u.drawWeapon(buffer, colors)
	}
}

const (
	uiColorFg = 15 // bright white
	uiColorBg = 0  // black
)

func (u *UILayer) drawCrosshair(buffer [][]rune, colors [][]raycast.ColorPair) {
	centerX := u.width / 2
	centerY := u.height / 2

	if centerY >= 0 && centerY < len(buffer) && centerX >= 0 && centerX < len(buffer[centerY]) {
		buffer[centerY][centerX] = '+'
		if colors != nil {
			colors[centerY][centerX] = raycast.ColorPair{Fg: uiColorFg, Bg: colors[centerY][centerX].Bg}
		}
	}
}

func (u *UILayer) drawHUD(buffer [][]rune, colors [][]raycast.ColorPair) {
	healthStr := fmt.Sprintf("HP:%3d", u.health)
	u.drawText(buffer, colors, 1, 0, healthStr)

	ammoStr := fmt.Sprintf("%2d/%2d", u.ammo, u.maxAmmo)
	u.drawText(buffer, colors, u.width-len(ammoStr)-1, 0, ammoStr)
}

var weaponSprite = []string{
	"    __    ",
	"   |__|   ",
	"  /    \\  ",
	" |      | ",
	" |______| ",
}

func (u *UILayer) drawWeapon(buffer [][]rune, colors [][]raycast.ColorPair) {
	spriteWidth := len(weaponSprite[0])
	spriteHeight := len(weaponSprite)

	startX := (u.width - spriteWidth) / 2
	startY := u.height - spriteHeight

	for dy, line := range weaponSprite {
		y := startY + dy
		if y < 0 || y >= len(buffer) {
			continue
		}
		for dx, ch := range line {
			x := startX + dx
			if x >= 0 && x < len(buffer[y]) && ch != ' ' {
				buffer[y][x] = ch
				if colors != nil {
					colors[y][x] = raycast.ColorPair{Fg: uiColorFg, Bg: colors[y][x].Bg}
				}
			}
		}
	}
}

func (u *UILayer) drawText(buffer [][]rune, colors [][]raycast.ColorPair, x, y int, text string) {
	if y < 0 || y >= len(buffer) {
		return
	}
	for i, ch := range text {
		px := x + i
		if px >= 0 && px < len(buffer[y]) {
			buffer[y][px] = ch
			if colors != nil {
				colors[y][px] = raycast.ColorPair{Fg: uiColorFg, Bg: colors[y][px].Bg}
			}
		}
	}
}

func (u *UILayer) OverlayToOutput(buf *bytes.Buffer, mainBuffer [][]rune, mainColors [][]raycast.ColorPair) {
	overlay := make([][]rune, len(mainBuffer))
	colors := make([][]raycast.ColorPair, len(mainBuffer))
	for i := range mainBuffer {
		overlay[i] = make([]rune, len(mainBuffer[i]))
		copy(overlay[i], mainBuffer[i])
		if mainColors != nil && i < len(mainColors) {
			colors[i] = make([]raycast.ColorPair, len(mainColors[i]))
			copy(colors[i], mainColors[i])
		}
	}

	u.Overlay(overlay, colors)

	for row := range overlay {
		for col := range overlay[row] {
			ch := overlay[row][col]
			cp := colors[row][col]
			buf.WriteString(fmt.Sprintf("\033[38;5;%dm\033[48;5;%dm%c", cp.Fg, cp.Bg, ch))
		}
		buf.WriteString("\033[0m\033[K\r\n")
	}
}
