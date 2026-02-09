package state

import (
	"bytes"
	"fmt"
	"log/slog"
	"time"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/terminal"
	"survival/internal/terminal/raycast"
	"survival/internal/terminal/session"
	"survival/internal/terminal/ui"
)

type SinglePlayerState struct {
	fd     int
	logger *slog.Logger

	session   *session.GameSession
	colliders []ports.Collider

	renderer25D  *raycast.Renderer25D
	uiLayer      *ui.UILayer
	viewHeight   float64
	currentInput ports.PlayerInput
}

func NewSinglePlayerState(fd int, logger *slog.Logger, sess *session.GameSession, colliders []ports.Collider) *SinglePlayerState {
	return &SinglePlayerState{
		fd:        fd,
		logger:    logger,
		session:   sess,
		colliders: colliders,
	}
}

func (s *SinglePlayerState) Init() {
	s.viewHeight = state.DefaultPlayerViewHeight

	viewWidth := terminal.AppDefaultConfig.Width
	viewHeight := terminal.AppDefaultConfig.Height - 2
	s.renderer25D = raycast.NewRenderer25D(viewWidth, viewHeight, s.viewHeight)
	s.uiLayer = ui.NewUILayer(viewWidth, viewHeight)
}

const fixedDeltaTime = 1.0 / 60.0

func (s *SinglePlayerState) Update(input terminal.InputEvent, _ time.Duration) terminal.Command {
	if input == terminal.InputCancel {
		return terminal.Command{Type: terminal.CmdPop}
	}

	s.handleGameInput(input)
	s.session.SetInput(s.currentInput)
	s.session.Update(fixedDeltaTime)

	return terminal.Command{Type: terminal.CmdNone}
}

func (s *SinglePlayerState) handleGameInput(input terminal.InputEvent) {
	switch input {
	case terminal.InputMoveForward:
		s.currentInput.MoveVertical = 1
	case terminal.InputMoveBackward:
		s.currentInput.MoveVertical = -1
	case terminal.InputMoveLeft:
		s.currentInput.MoveHorizontal = -1
	case terminal.InputMoveRight:
		s.currentInput.MoveHorizontal = 1
	case terminal.InputTurnLeft:
		s.currentInput.LookHorizontal = -1
	case terminal.InputTurnRight:
		s.currentInput.LookHorizontal = 1
	case terminal.InputNone:
		s.currentInput = ports.PlayerInput{}
	}

	s.currentInput.MovementType = ports.MovementTypeRelative
}

func (s *SinglePlayerState) Draw(buf *bytes.Buffer, width, height int) {
	x, y, dir, ok := s.session.PlayerState()
	if !ok {
		s.drawCenteredMessage(buf, width, height, "Loading...")
		return
	}

	s.drawGameView(buf, width, height, x, y, dir)
}

func (s *SinglePlayerState) drawCenteredMessage(buf *bytes.Buffer, width, height int, message string) {
	buf.WriteString(setGreenFont)

	for i := 0; i < height/2-1; i++ {
		buf.WriteString("\033[K\r\n")
	}

	drawCenteredLine(buf, width, message)

	for i := height/2 + 1; i < height; i++ {
		buf.WriteString("\033[K\r\n")
	}

	buf.WriteString(resetFontColor)
}

func (s *SinglePlayerState) drawGameView(buf *bytes.Buffer, width, height int, playerX, playerY, playerDir float64) {
	colliders := s.colliders

	numRays := width
	results := raycast.CastRays(playerX, playerY, playerDir, s.viewHeight, colliders, numRays)

	s.renderer25D.Render(results)

	outputBuf := s.renderer25D.GetOutputBuffer()
	colorBuf := s.renderer25D.GetColorBuffer()
	s.uiLayer.Overlay(outputBuf, colorBuf)

	s.renderer25D.WriteWithOverlay(buf)

	locale := terminal.AppDefaultConfig.Locale
	statusLine := fmt.Sprintf("X:%.1f Y:%.1f Dir:%.2f | %s", playerX, playerY, playerDir, locale.SPStatusHint)
	drawCenteredLine(buf, width, statusLine)
}
