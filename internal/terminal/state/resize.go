package state

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/term"

	"survival/internal/terminal"
)

const (
	setRedFont     = "\033[31m"
	setGreenFont   = "\033[32m"
	resetFontColor = "\033[0m"
)

// ResizeState resize current terminal state
type ResizeState struct {
	fd       int
	isReady  bool
	currentW int
	currentH int
	logger   *slog.Logger
}

func NewResizeState(fd int, logger *slog.Logger) *ResizeState {
	return &ResizeState{
		fd:     fd,
		logger: logger,
	}
}

func (s *ResizeState) Init() {
	s.checkSize()
}

func (s *ResizeState) checkSize() {
	width, height, err := term.GetSize(s.fd)
	if err != nil {
		return
	}
	s.currentW = width
	s.currentH = height
	s.isReady = width >= terminal.AppDefaultConfig.Width && height >= terminal.AppDefaultConfig.Height
}

func (s *ResizeState) Update(input terminal.InputEvent, dt time.Duration) terminal.Command {
	s.checkSize()

	switch input {
	case terminal.InputCancel:
		return terminal.Command{Type: terminal.CmdQuit}

	case terminal.InputAction:
		if s.isReady {
			return terminal.Command{Type: terminal.CmdSwap, NextState: NewMainMenuState(s.fd, s.logger)}
		}
	default:
		s.logger.Debug("ResizeState received unhandled input")
	}

	return terminal.Command{Type: terminal.CmdNone}
}

func (s *ResizeState) Draw(buf *bytes.Buffer, width, height int) {
	locale := terminal.AppDefaultConfig.Locale
	targetW := terminal.AppDefaultConfig.Width
	targetH := terminal.AppDefaultConfig.Height

	var title, msg string
	colorCode := setRedFont
	if s.isReady {
		colorCode = setGreenFont
		title = locale.TitleReady
		msg = locale.MsgReady
	} else {
		title = locale.TitleWait
		msg = locale.MsgWait
	}

	boxWidth := width
	boxHeight := height
	leftPadding := 0
	topEmptyRows := 0

	if width > targetW && height > targetH {
		boxWidth = targetW
		boxHeight = targetH
		leftPadding = (width - targetW) / 2
		topEmptyRows = (height - targetH) / 2
	}

	borderLine := strings.Repeat(locale.BoxBorderH, boxWidth-2)
	borderTop := fmt.Sprintf("╔%s╗", borderLine)
	borderBot := fmt.Sprintf("╚%s╝", borderLine)
	emptyRow := DrawBoxRow("", boxWidth, locale)

	contentRows := []string{
		DrawBoxRow(title, boxWidth, locale),
		emptyRow,
		FormatRow(locale.LblCurrent, fmt.Sprintf("%d x %d", s.currentW, s.currentH), boxWidth, locale),
		FormatRow(locale.LblTarget, fmt.Sprintf("%d x %d", targetW, targetH), boxWidth, locale),
		emptyRow,
		DrawBoxRow(msg, boxWidth, locale),
	}

	innerHeight := boxHeight - 2
	contentHeight := len(contentRows)
	topPadding := (innerHeight - contentHeight) / 2

	buf.WriteString(colorCode)

	padLeft := strings.Repeat(" ", leftPadding)

	for i := 0; i < topEmptyRows; i++ {
		writeLine(buf, "")
	}

	writeLine(buf, padLeft+borderTop)

	for i := 0; i < topPadding; i++ {
		writeLine(buf, padLeft+emptyRow)
	}

	for _, row := range contentRows {
		writeLine(buf, padLeft+row)
	}

	bottomPadding := innerHeight - topPadding - contentHeight
	for i := 0; i < bottomPadding; i++ {
		writeLine(buf, padLeft+emptyRow)
	}

	writeLine(buf, padLeft+borderBot)

	buf.WriteString(resetFontColor)
}

func drawCenteredLine(buf *bytes.Buffer, screenWidth int, content string) {
	padding := (screenWidth - visibleLength(content)) / 2
	if padding > 0 {
		buf.WriteString(strings.Repeat(" ", padding))
	}
	buf.WriteString(content)
	buf.WriteString("\033[K\r\n")
}

func writeLine(buf *bytes.Buffer, content string) {
	buf.WriteString(content)
	buf.WriteString("\033[K\r\n")
}

func DrawBoxRow(text string, width int, loc terminal.LocaleData) string {
	contentWidth := width - 2 // exclude borders
	paddedText := PadCenter(text, contentWidth)
	return fmt.Sprintf("%s%s%s", loc.BoxBorderV, paddedText, loc.BoxBorderV)
}

func FormatRow(label, value string, width int, loc terminal.LocaleData) string {
	contentWidth := width - 4

	availableSpace := contentWidth - visibleLength(label) - visibleLength(value)
	if availableSpace < 0 {
		availableSpace = 0
	}
	space := strings.Repeat(" ", availableSpace)
	rowContent := fmt.Sprintf(" %s%s%s ", label, space, value)
	return fmt.Sprintf("%s%s%s", loc.BoxBorderV, rowContent, loc.BoxBorderV)
}

func PadCenter(s string, width int) string {
	l := visibleLength(s)
	if l >= width {
		return s
	}
	leftPad := (width - l) / 2
	rightPad := width - l - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

func visibleLength(s string) int {
	length := 0
	for _, r := range s {
		if isDoubleWidth(r) {
			length += 2
		} else {
			length += 1
		}
	}
	return length
}

func isDoubleWidth(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFF00 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6) ||
		(r >= 0x20000 && r <= 0x2A6DF) ||
		(r >= 0x2A700 && r <= 0x2B73F) ||
		(r >= 0x2B740 && r <= 0x2B81F) ||
		(r >= 0x2B820 && r <= 0x2CEAF) ||
		(r >= 0x2CEB0 && r <= 0x2EBEF) ||
		(r >= 0x30000 && r <= 0x3134F) ||
		(r >= 0x3100 && r <= 0x312F) ||
		(r >= 0x31A0 && r <= 0x31BF) ||
		(r >= 0xAC00 && r <= 0xD7AF) ||
		(r >= 0x1100 && r <= 0x11FF)
}
