package state

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"survival/internal/terminal"
)

type MainMenuState struct {
	fd            int
	selectedIndex int
	logger        *slog.Logger
}

func NewMainMenuState(fd int, logger *slog.Logger) *MainMenuState {
	return &MainMenuState{
		fd:            fd,
		selectedIndex: 0,
		logger:        logger,
	}
}

func (s *MainMenuState) Init() {
	s.selectedIndex = 0
}

func (s *MainMenuState) Update(input terminal.InputEvent, dt time.Duration) terminal.Command {
	menuItemCount := 4

	switch input {
	case terminal.InputMoveBackward:
		s.selectedIndex--
		if s.selectedIndex < 0 {
			s.selectedIndex = menuItemCount - 1
		}

	case terminal.InputMoveForward:
		s.selectedIndex++
		if s.selectedIndex >= menuItemCount {
			s.selectedIndex = 0
		}

	case terminal.InputAction:
		switch s.selectedIndex {
		case 0, 1:
			return terminal.Command{Type: terminal.CmdNone}
		case 2:
			return terminal.Command{Type: terminal.CmdPush, NextState: NewSettingState(s.fd, s.logger)}
		case 3:
			return terminal.Command{Type: terminal.CmdQuit}
		}

	case terminal.InputCancel:
		return terminal.Command{Type: terminal.CmdQuit}
	}

	return terminal.Command{Type: terminal.CmdNone}
}

func (s *MainMenuState) Draw(buf *bytes.Buffer, width, height int) {
	locale := terminal.AppDefaultConfig.Locale
	boxWidth := 50

	menuItems := []string{
		locale.MenuStart,
		locale.MenuMulti,
		locale.MenuSettings,
		locale.MenuExit,
	}

	borderLine := strings.Repeat(locale.BoxBorderH, boxWidth-2)
	borderTop := fmt.Sprintf("╔%s╗", borderLine)
	borderBot := fmt.Sprintf("╚%s╝", borderLine)
	emptyRow := DrawBoxRow("", boxWidth, locale)

	buf.WriteString(setGreenFont)

	drawCenteredLine(buf, width, borderTop)
	drawCenteredLine(buf, width, emptyRow)
	drawCenteredLine(buf, width, DrawBoxRow(locale.MenuTitle, boxWidth, locale))
	drawCenteredLine(buf, width, emptyRow)

	for i, item := range menuItems {
		var indicator string
		if i == s.selectedIndex {
			indicator = "► "
		} else {
			indicator = "  "
		}
		rowText := indicator + item
		drawCenteredLine(buf, width, DrawBoxRow(rowText, boxWidth, locale))
	}

	drawCenteredLine(buf, width, emptyRow)
	drawCenteredLine(buf, width, borderBot)

	buf.WriteString(resetFontColor)
	drawCenteredLine(buf, width, "")
	drawCenteredLine(buf, width, PadCenter(locale.MenuHint, boxWidth))
}
