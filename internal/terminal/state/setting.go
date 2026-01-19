package state

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"survival/internal/terminal"
)

var screenSizes = []struct {
	Width  int
	Height int
}{
	{120, 32},
	{160, 40},
	{200, 50},
}

type SettingState struct {
	fd            int
	selectedIndex int
	sizeIndex     int
	logger        *slog.Logger
}

func NewSettingState(fd int, logger *slog.Logger) *SettingState {
	sizeIdx := 0
	for i, size := range screenSizes {
		if size.Width == terminal.AppDefaultConfig.Width && size.Height == terminal.AppDefaultConfig.Height {
			sizeIdx = i
			break
		}
	}

	return &SettingState{
		fd:            fd,
		selectedIndex: 0,
		sizeIndex:     sizeIdx,
		logger:        logger,
	}
}

func (s *SettingState) Init() {
	s.selectedIndex = 0
}

func (s *SettingState) Update(input terminal.InputEvent, dt time.Duration) terminal.Command {
	settingItemCount := 2

	switch input {
	case terminal.InputMoveBackward:
		s.selectedIndex--
		if s.selectedIndex < 0 {
			s.selectedIndex = settingItemCount - 1
		}

	case terminal.InputMoveForward:
		s.selectedIndex++
		if s.selectedIndex >= settingItemCount {
			s.selectedIndex = 0
		}

	case terminal.InputAction:
		switch s.selectedIndex {
		case 0:
			s.sizeIndex++
			if s.sizeIndex >= len(screenSizes) {
				s.sizeIndex = 0
			}
			terminal.AppDefaultConfig.Width = screenSizes[s.sizeIndex].Width
			terminal.AppDefaultConfig.Height = screenSizes[s.sizeIndex].Height
			return terminal.Command{Type: terminal.CmdSwap, NextState: NewResizeState(s.fd, s.logger)}

		case 1:
			if terminal.AppDefaultConfig.Locale.TitleReady == terminal.LangTW.TitleReady {
				terminal.AppDefaultConfig.Locale = terminal.LangEN
			} else {
				terminal.AppDefaultConfig.Locale = terminal.LangTW
			}
		}

	case terminal.InputCancel:
		return terminal.Command{Type: terminal.CmdPop}
	}

	return terminal.Command{Type: terminal.CmdNone}
}

func (s *SettingState) Draw(buf *bytes.Buffer, width, height int) {
	locale := terminal.AppDefaultConfig.Locale
	boxWidth := 60

	var sizeLabel string
	switch s.sizeIndex {
	case 0:
		sizeLabel = locale.SizeSmall
	case 1:
		sizeLabel = locale.SizeMedium
	case 2:
		sizeLabel = locale.SizeLarge
	}

	var langLabel string
	if terminal.AppDefaultConfig.Locale.TitleReady == terminal.LangTW.TitleReady {
		langLabel = locale.LangNameTW
	} else {
		langLabel = locale.LangNameEN
	}

	settingItems := []struct {
		label string
		value string
	}{
		{locale.SettingScreenSize, sizeLabel},
		{locale.SettingLanguage, langLabel},
	}

	borderLine := strings.Repeat(locale.BoxBorderH, boxWidth-2)
	borderTop := fmt.Sprintf("╔%s╗", borderLine)
	borderBot := fmt.Sprintf("╚%s╝", borderLine)
	emptyRow := DrawBoxRow("", boxWidth, locale)

	buf.WriteString(setGreenFont)

	drawCenteredLine(buf, width, borderTop)
	drawCenteredLine(buf, width, emptyRow)
	drawCenteredLine(buf, width, DrawBoxRow(locale.SettingsTitle, boxWidth, locale))
	drawCenteredLine(buf, width, emptyRow)

	for i, item := range settingItems {
		var indicator string
		if i == s.selectedIndex {
			indicator = "► "
		} else {
			indicator = "  "
		}
		rowText := fmt.Sprintf("%s%-16s: %s", indicator, item.label, item.value)
		drawCenteredLine(buf, width, DrawBoxRow(rowText, boxWidth, locale))
	}

	drawCenteredLine(buf, width, emptyRow)
	drawCenteredLine(buf, width, borderBot)

	buf.WriteString(resetFontColor)
	drawCenteredLine(buf, width, "")
	drawCenteredLine(buf, width, PadCenter(locale.SettingsHint, boxWidth))
}
