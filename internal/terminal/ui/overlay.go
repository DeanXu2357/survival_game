package ui

import (
	"bytes"
	"fmt"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/terminal/raycast"
)

type WeaponSlotInfo struct {
	Type      state.WeaponType
	Name      string
	Occupied  bool
	Ammo      int
	MaxAmmo   int
	SpareMags int
}

type UILayer struct {
	width               int
	height              int
	crosshairEnabled    bool
	weaponEnabled       bool
	hudEnabled          bool
	health              int
	weaponSlots         [3]WeaponSlotInfo
	currentWeaponIndex  int
	currentWeaponType   state.WeaponType
	currentTick         ports.Tick
	lastFireTick        ports.Tick
	reloadStartTick     ports.Tick
	reloadDurationTicks ports.Tick
}

func NewUILayer(width, height int) *UILayer {
	return &UILayer{
		width:            width,
		height:           height,
		crosshairEnabled: true,
		weaponEnabled:    true,
		hudEnabled:       true,
		health:           100,
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

func (u *UILayer) SetWeaponSlots(slots [3]WeaponSlotInfo, currentIndex int) {
	u.weaponSlots = slots
	u.currentWeaponIndex = currentIndex
	if currentIndex >= 0 && currentIndex < len(slots) {
		u.currentWeaponType = slots[currentIndex].Type
	}
}

func (u *UILayer) SetFireState(currentTick, lastFireTick ports.Tick) {
	u.currentTick = currentTick
	u.lastFireTick = lastFireTick
}

func (u *UILayer) SetReloadState(reloadStartTick, reloadDurationTicks ports.Tick) {
	u.reloadStartTick = reloadStartTick
	u.reloadDurationTicks = reloadDurationTicks
}

func (u *UILayer) Overlay(buffer [][]rune, colors [][]raycast.ColorPair) {
	if u.crosshairEnabled {
		u.drawCrosshair(buffer, colors)
	}

	if u.hudEnabled {
		u.drawHUD(buffer, colors)
		u.drawWeaponPanel(buffer, colors)
	}

	if u.weaponEnabled {
		u.drawBulletTrail(buffer, colors)
		u.drawWeapon(buffer, colors)
		u.drawMuzzleFlash(buffer, colors)
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
}

const weaponPanelWidth = 22

func (u *UILayer) drawWeaponPanel(buffer [][]rune, colors [][]raycast.ColorPair) {
	var lines []string
	for i, slot := range u.weaponSlots {
		if !slot.Occupied {
			continue
		}
		prefix := "  "
		if i == u.currentWeaponIndex {
			prefix = "► "
		}
		if slot.MaxAmmo > 0 {
			ammo := fmt.Sprintf("%2d/%2d [%d]", slot.Ammo, slot.MaxAmmo, slot.SpareMags)
			pad := weaponPanelWidth - 4 - len(prefix) - len([]rune(slot.Name)) - len(ammo)
			if pad < 1 {
				pad = 1
			}
			line := prefix + slot.Name
			for j := 0; j < pad; j++ {
				line += " "
			}
			line += ammo
			lines = append(lines, line)
		} else {
			line := prefix + slot.Name
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		return
	}

	panelHeight := len(lines) + 2
	innerWidth := weaponPanelWidth - 2
	startX := u.width - weaponPanelWidth - 1
	startY := u.height - panelHeight

	top := "┌" + repeatRune('─', innerWidth) + "┐"
	bottom := "└" + repeatRune('─', innerWidth) + "┘"

	u.drawText(buffer, colors, startX, startY, top)
	for i, line := range lines {
		padded := padRight(line, innerWidth)
		row := "│" + padded + "│"
		u.drawText(buffer, colors, startX, startY+1+i, row)
	}
	u.drawText(buffer, colors, startX, startY+panelHeight-1, bottom)
}

func repeatRune(r rune, n int) string {
	out := make([]rune, n)
	for i := range out {
		out[i] = r
	}
	return string(out)
}

func padRight(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return string(runes[:width])
	}
	pad := make([]rune, width-len(runes))
	for i := range pad {
		pad[i] = ' '
	}
	return s + string(pad)
}

var weaponSprites = map[state.WeaponType][]string{
	state.WeaponTypeFist: {
		"    ___   ",
		"   | _ |  ",
		"   || ||  ",
		"   || ||  ",
		"   |___|  ",
	},
	state.WeaponTypeKnife: {
		"     |    ",
		"     |    ",
		"    /|\\   ",
		"   |___|  ",
		"    | |   ",
	},
	state.WeaponTypeGun: {
		"    __    ",
		"   |__|   ",
		"  /    \\  ",
		" |      | ",
		" |______| ",
	},
}

var weaponAttackSprites = map[state.WeaponType][]string{
	state.WeaponTypeFist: {
		"   ___    ",
		"  |   |   ",
		"  | _ |   ",
		"  || ||   ",
		"  || ||   ",
		"  || ||   ",
		"  |___|   ",
	},
	state.WeaponTypeKnife: {
		"     |    ",
		"     |    ",
		"     |    ",
		"     |    ",
		"    /|\\   ",
		"   |___|  ",
		"    | |   ",
	},
}

const (
	recoilDuration   ports.Tick = 10
	recoilRiseTicks  ports.Tick = 3
	recoilPeakOffset            = 3

	meleeAttackDuration ports.Tick = 14
	meleeRiseTicks      ports.Tick = 4
	fistPeakYOffset                = 4
	knifePeakYOffset               = 5

	fistRestingXOffset  = 10
	knifeRestingXOffset = 6

	reloadDropRows = 5
)

func recoilOffset(elapsed ports.Tick) int {
	if elapsed >= recoilDuration {
		return 0
	}
	if elapsed <= recoilRiseTicks {
		return int(elapsed) * recoilPeakOffset / int(recoilRiseTicks)
	}
	remaining := recoilDuration - elapsed
	fallTicks := recoilDuration - recoilRiseTicks
	return int(remaining) * recoilPeakOffset / int(fallTicks)
}

func meleeXOffset(elapsed, duration, riseTicks ports.Tick, restingOffset int) int {
	if elapsed >= duration {
		return restingOffset
	}
	if elapsed <= riseTicks {
		return restingOffset - int(elapsed)*restingOffset/int(riseTicks)
	}
	remaining := duration - elapsed
	fallTicks := duration - riseTicks
	return restingOffset - int(remaining)*restingOffset/int(fallTicks)
}

func reloadYOffset(elapsed, duration ports.Tick) int {
	if elapsed >= duration {
		return 0
	}
	progress := float64(elapsed) / float64(duration)
	switch {
	case progress < 0.25:
		return int(progress / 0.25 * float64(reloadDropRows))
	case progress < 0.75:
		return reloadDropRows
	default:
		return int((1.0 - progress) / 0.25 * float64(reloadDropRows))
	}
}

func meleeYOffset(elapsed, duration, riseTicks ports.Tick, peakOffset int) int {
	if elapsed >= duration {
		return 0
	}
	if elapsed <= riseTicks {
		return int(elapsed) * peakOffset / int(riseTicks)
	}
	remaining := duration - elapsed
	fallTicks := duration - riseTicks
	return int(remaining) * peakOffset / int(fallTicks)
}

const (
	muzzleFlashDuration ports.Tick = 4
	bulletTrailDuration ports.Tick = 3
)

var (
	muzzleFlashColors = [4]int{226, 208, 172, 172}
	bulletTrailColors = [3]int{231, 245, 240}
)

func (u *UILayer) gunFireElapsed() (ports.Tick, bool) {
	if u.currentWeaponType != state.WeaponTypeGun {
		return 0, false
	}
	if u.lastFireTick == 0 || u.currentTick < u.lastFireTick {
		return 0, false
	}
	isReloading := u.reloadStartTick > 0 && u.currentTick >= u.reloadStartTick &&
		(u.currentTick-u.reloadStartTick) < u.reloadDurationTicks
	if isReloading {
		return 0, false
	}
	return u.currentTick - u.lastFireTick, true
}

func (u *UILayer) drawMuzzleFlash(buffer [][]rune, colors [][]raycast.ColorPair) {
	elapsed, active := u.gunFireElapsed()
	if !active || elapsed >= muzzleFlashDuration {
		return
	}

	sprite := weaponSprites[state.WeaponTypeGun]
	spriteHeight := len(sprite)
	baseY := u.height - spriteHeight - recoilOffset(elapsed)
	centerX := u.width / 2

	fg := muzzleFlashColors[elapsed]

	type flashCell struct {
		dx, dy int
		ch     rune
	}

	var pattern []flashCell
	if elapsed <= 1 {
		pattern = []flashCell{
			{-2, -2, '·'}, {0, -2, '*'}, {2, -2, '·'},
			{-1, -1, '*'}, {0, -1, '*'}, {1, -1, '*'},
			{0, 0, '*'},
		}
	} else {
		pattern = []flashCell{
			{0, -1, '*'},
			{0, 0, '*'},
		}
	}

	for _, p := range pattern {
		x := centerX + p.dx
		y := baseY + p.dy
		if y >= 0 && y < len(buffer) && x >= 0 && x < len(buffer[y]) {
			buffer[y][x] = p.ch
			if colors != nil {
				colors[y][x] = raycast.ColorPair{Fg: fg, Bg: colors[y][x].Bg}
			}
		}
	}
}

func (u *UILayer) drawBulletTrail(buffer [][]rune, colors [][]raycast.ColorPair) {
	elapsed, active := u.gunFireElapsed()
	if !active || elapsed >= bulletTrailDuration {
		return
	}

	sprite := weaponSprites[state.WeaponTypeGun]
	spriteHeight := len(sprite)
	gunTopY := u.height - spriteHeight - recoilOffset(elapsed)
	crosshairY := u.height / 2
	centerX := u.width / 2

	fg := bulletTrailColors[elapsed]

	for y := crosshairY + 1; y < gunTopY-2; y++ {
		if y >= 0 && y < len(buffer) && centerX >= 0 && centerX < len(buffer[y]) {
			buffer[y][centerX] = '│'
			if colors != nil {
				colors[y][centerX] = raycast.ColorPair{Fg: fg, Bg: colors[y][centerX].Bg}
			}
		}
	}
}

func (u *UILayer) drawWeapon(buffer [][]rune, colors [][]raycast.ColorPair) {
	sprite, ok := weaponSprites[u.currentWeaponType]
	if !ok {
		sprite = weaponSprites[state.WeaponTypeFist]
	}

	isReloading := u.reloadStartTick > 0 && u.currentTick >= u.reloadStartTick &&
		(u.currentTick-u.reloadStartTick) < u.reloadDurationTicks
	if !isReloading && u.lastFireTick > 0 && u.currentTick >= u.lastFireTick {
		elapsed := u.currentTick - u.lastFireTick
		if elapsed < meleeAttackDuration {
			if attackSprite, hasAttack := weaponAttackSprites[u.currentWeaponType]; hasAttack {
				sprite = attackSprite
			}
		}
	}

	spriteWidth := len(sprite[0])
	spriteHeight := len(sprite)

	startX := (u.width - spriteWidth) / 2
	startY := u.height - spriteHeight

	if isReloading {
		startY += reloadYOffset(u.currentTick-u.reloadStartTick, u.reloadDurationTicks)
		switch u.currentWeaponType {
		case state.WeaponTypeFist:
			startX += fistRestingXOffset
		case state.WeaponTypeKnife:
			startX += knifeRestingXOffset
		}
	} else if u.lastFireTick > 0 && u.currentTick >= u.lastFireTick {
		elapsed := u.currentTick - u.lastFireTick
		switch u.currentWeaponType {
		case state.WeaponTypeFist:
			startX += meleeXOffset(elapsed, meleeAttackDuration, meleeRiseTicks, fistRestingXOffset)
			startY -= meleeYOffset(elapsed, meleeAttackDuration, meleeRiseTicks, fistPeakYOffset)
		case state.WeaponTypeKnife:
			startX += meleeXOffset(elapsed, meleeAttackDuration, meleeRiseTicks, knifeRestingXOffset)
			startY -= meleeYOffset(elapsed, meleeAttackDuration, meleeRiseTicks, knifePeakYOffset)
		default:
			startY -= recoilOffset(elapsed)
		}
	} else {
		switch u.currentWeaponType {
		case state.WeaponTypeFist:
			startX += fistRestingXOffset
		case state.WeaponTypeKnife:
			startX += knifeRestingXOffset
		}
	}

	for dy, line := range sprite {
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
