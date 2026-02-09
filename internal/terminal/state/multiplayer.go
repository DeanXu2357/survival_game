package state

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/terminal"
	"survival/internal/terminal/network"
	"survival/internal/terminal/raycast"
	"survival/internal/terminal/ui"
)

type MultiplayerPhase int

const (
	MultiplayerPhaseConnecting MultiplayerPhase = iota
	MultiplayerPhaseWaitingForRoom
	MultiplayerPhaseJoiningRoom
	MultiplayerPhasePlaying
	MultiplayerPhaseError
)

const (
	multiplayerServerAddr = "localhost:3033"
)

type MultiplayerState struct {
	fd     int
	logger *slog.Logger

	phase    MultiplayerPhase
	client   *network.Client
	clientID string

	playerX   float64
	playerY   float64
	playerDir float64
	colliders []ports.Collider

	renderer25D  *raycast.Renderer25D
	uiLayer      *ui.UILayer
	viewHeight   float64
	errorMessage string

	currentInput ports.PlayerInput
	inputChanged bool
}

func NewMultiplayerState(fd int, logger *slog.Logger) *MultiplayerState {
	return &MultiplayerState{
		fd:       fd,
		logger:   logger,
		clientID: generateMultiplayerClientID(),
	}
}

func generateMultiplayerClientID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "term-" + hex.EncodeToString(b)
}

func (s *MultiplayerState) Init() {
	s.phase = MultiplayerPhaseConnecting
	s.client = network.NewClient(s.clientID)
	s.viewHeight = state.DefaultPlayerViewHeight

	viewWidth := terminal.AppDefaultConfig.Width
	viewHeight := terminal.AppDefaultConfig.Height - 2
	s.renderer25D = raycast.NewRenderer25D(viewWidth, viewHeight, s.viewHeight)
	s.uiLayer = ui.NewUILayer(viewWidth, viewHeight)

	go s.connectAndJoin()
}

func (s *MultiplayerState) connectAndJoin() {
	err := s.client.Connect(multiplayerServerAddr, "Player")
	if err != nil {
		s.logger.Error("Failed to connect", "error", err)
		s.errorMessage = fmt.Sprintf("Connection failed: %v", err)
		s.phase = MultiplayerPhaseError
		return
	}

	s.phase = MultiplayerPhaseWaitingForRoom
	err = s.client.RequestRoomList()
	if err != nil {
		s.logger.Error("Failed to request room list", "error", err)
		s.errorMessage = fmt.Sprintf("Failed to get rooms: %v", err)
		s.phase = MultiplayerPhaseError
		return
	}
}

func (s *MultiplayerState) Update(input terminal.InputEvent, dt time.Duration) terminal.Command {
	s.processNetworkMessages()

	switch s.phase {
	case MultiplayerPhaseError:
		if input == terminal.InputCancel || input == terminal.InputAction {
			s.cleanup()
			return terminal.Command{Type: terminal.CmdPop}
		}

	case MultiplayerPhasePlaying:
		if input == terminal.InputCancel {
			s.cleanup()
			return terminal.Command{Type: terminal.CmdPop}
		}
		s.handleGameInput(input)
		s.sendInputIfChanged()
	}

	return terminal.Command{Type: terminal.CmdNone}
}

func (s *MultiplayerState) processNetworkMessages() {
	for {
		select {
		case roomList := <-s.client.RoomListChan():
			s.handleRoomList(roomList)
		case <-s.client.JoinSuccessChan():
			s.phase = MultiplayerPhasePlaying
			s.logger.Info("Joined room successfully")
		case update := <-s.client.GameUpdateChan():
			s.handleGameUpdate(update)
		case staticData := <-s.client.StaticDataChan():
			s.handleStaticData(staticData)
		case err := <-s.client.ErrorChan():
			s.logger.Error("Network error", "error", err)
			s.errorMessage = err.Error()
			s.phase = MultiplayerPhaseError
		default:
			return
		}
	}
}

func (s *MultiplayerState) handleRoomList(roomList ports.ListRoomsResponse) {
	if len(roomList.Rooms) == 0 {
		s.errorMessage = "No rooms available"
		s.phase = MultiplayerPhaseError
		return
	}

	s.phase = MultiplayerPhaseJoiningRoom
	err := s.client.RequestJoinRoom(roomList.Rooms[0].RoomID)
	if err != nil {
		s.errorMessage = fmt.Sprintf("Failed to join room: %v", err)
		s.phase = MultiplayerPhaseError
	}
}

func (s *MultiplayerState) handleGameUpdate(update ports.GameUpdatePayload) {
	s.playerX = update.Me.X
	s.playerY = update.Me.Y
	s.playerDir = update.Me.Dir
}

func (s *MultiplayerState) handleStaticData(data ports.StaticDataPayload) {
	s.colliders = data.Colliders
	s.logger.Info("Received static data", "colliders", len(s.colliders))
}

func (s *MultiplayerState) handleGameInput(input terminal.InputEvent) {
	prevInput := s.currentInput

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

	if s.currentInput != prevInput {
		s.inputChanged = true
	}
}

func (s *MultiplayerState) sendInputIfChanged() {
	if !s.inputChanged {
		return
	}

	err := s.client.SendInput(s.currentInput)
	if err != nil {
		s.logger.Error("Failed to send input", "error", err)
	}
	s.inputChanged = false
}

func (s *MultiplayerState) cleanup() {
	if s.client != nil {
		s.client.Close()
	}
}

func (s *MultiplayerState) Draw(buf *bytes.Buffer, width, height int) {
	locale := terminal.AppDefaultConfig.Locale

	switch s.phase {
	case MultiplayerPhaseConnecting:
		s.drawCenteredMessage(buf, width, height, locale.SPConnecting)
	case MultiplayerPhaseWaitingForRoom:
		s.drawCenteredMessage(buf, width, height, locale.SPWaitingRoom)
	case MultiplayerPhaseJoiningRoom:
		s.drawCenteredMessage(buf, width, height, locale.SPJoiningRoom)
	case MultiplayerPhaseError:
		s.drawErrorScreen(buf, width, height)
	case MultiplayerPhasePlaying:
		s.drawGameView(buf, width, height)
	}
}

func (s *MultiplayerState) drawCenteredMessage(buf *bytes.Buffer, width, height int, message string) {
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

func (s *MultiplayerState) drawErrorScreen(buf *bytes.Buffer, width, height int) {
	locale := terminal.AppDefaultConfig.Locale

	buf.WriteString(setRedFont)

	for i := 0; i < height/2-2; i++ {
		buf.WriteString("\033[K\r\n")
	}

	drawCenteredLine(buf, width, locale.SPError)
	drawCenteredLine(buf, width, s.errorMessage)
	drawCenteredLine(buf, width, "")
	drawCenteredLine(buf, width, locale.SPStatusHint)

	for i := height/2 + 2; i < height; i++ {
		buf.WriteString("\033[K\r\n")
	}

	buf.WriteString(resetFontColor)
}

func (s *MultiplayerState) drawGameView(buf *bytes.Buffer, width, height int) {
	playerX := s.playerX
	playerY := s.playerY
	playerDir := s.playerDir
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
