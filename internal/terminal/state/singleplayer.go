package state

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"survival/internal/core/ports"
	"survival/internal/terminal"
	"survival/internal/terminal/network"
	"survival/internal/terminal/raycast"
)

type GamePhase int

const (
	PhaseConnecting GamePhase = iota
	PhaseWaitingForRoom
	PhaseJoiningRoom
	PhasePlaying
	PhaseError
)

const (
	serverAddr = "localhost:3033"
)

type SinglePlayerState struct {
	fd     int
	logger *slog.Logger

	phase    GamePhase
	client   *network.Client
	clientID string

	playerX   float64
	playerY   float64
	playerDir float64
	colliders []ports.Collider
	dataMu    sync.RWMutex

	renderer     *raycast.Renderer
	errorMessage string

	currentInput ports.PlayerInput
	inputChanged bool
}

func NewSinglePlayerState(fd int, logger *slog.Logger) *SinglePlayerState {
	return &SinglePlayerState{
		fd:       fd,
		logger:   logger,
		clientID: generateClientID(),
	}
}

func generateClientID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "term-" + hex.EncodeToString(b)
}

func (s *SinglePlayerState) Init() {
	s.phase = PhaseConnecting
	s.client = network.NewClient(s.clientID)
	s.renderer = raycast.NewRenderer(terminal.AppDefaultConfig.Width, terminal.AppDefaultConfig.Height-2)

	go s.connectAndJoin()
}

func (s *SinglePlayerState) connectAndJoin() {
	err := s.client.Connect(serverAddr, "Player")
	if err != nil {
		s.logger.Error("Failed to connect", "error", err)
		s.errorMessage = fmt.Sprintf("Connection failed: %v", err)
		s.phase = PhaseError
		return
	}

	s.phase = PhaseWaitingForRoom
	err = s.client.RequestRoomList()
	if err != nil {
		s.logger.Error("Failed to request room list", "error", err)
		s.errorMessage = fmt.Sprintf("Failed to get rooms: %v", err)
		s.phase = PhaseError
		return
	}
}

func (s *SinglePlayerState) Update(input terminal.InputEvent, dt time.Duration) terminal.Command {
	s.processNetworkMessages()

	switch s.phase {
	case PhaseError:
		if input == terminal.InputCancel || input == terminal.InputAction {
			s.cleanup()
			return terminal.Command{Type: terminal.CmdPop}
		}

	case PhasePlaying:
		if input == terminal.InputCancel {
			s.cleanup()
			return terminal.Command{Type: terminal.CmdPop}
		}
		s.handleGameInput(input)
		s.sendInputIfChanged()
	}

	return terminal.Command{Type: terminal.CmdNone}
}

func (s *SinglePlayerState) processNetworkMessages() {
	for {
		select {
		case roomList := <-s.client.RoomListChan():
			s.handleRoomList(roomList)
		case <-s.client.JoinSuccessChan():
			s.phase = PhasePlaying
			s.logger.Info("Joined room successfully")
		case update := <-s.client.GameUpdateChan():
			s.handleGameUpdate(update)
		case staticData := <-s.client.StaticDataChan():
			s.handleStaticData(staticData)
		case err := <-s.client.ErrorChan():
			s.logger.Error("Network error", "error", err)
			s.errorMessage = err.Error()
			s.phase = PhaseError
		default:
			return
		}
	}
}

func (s *SinglePlayerState) handleRoomList(roomList ports.ListRoomsResponse) {
	if len(roomList.Rooms) == 0 {
		s.errorMessage = "No rooms available"
		s.phase = PhaseError
		return
	}

	s.phase = PhaseJoiningRoom
	err := s.client.RequestJoinRoom(roomList.Rooms[0].RoomID)
	if err != nil {
		s.errorMessage = fmt.Sprintf("Failed to join room: %v", err)
		s.phase = PhaseError
	}
}

func (s *SinglePlayerState) handleGameUpdate(update ports.GameUpdatePayload) {
	s.dataMu.Lock()
	defer s.dataMu.Unlock()

	s.playerX = update.Me.X
	s.playerY = update.Me.Y
	s.playerDir = update.Me.Dir
}

func (s *SinglePlayerState) handleStaticData(data ports.StaticDataPayload) {
	s.dataMu.Lock()
	defer s.dataMu.Unlock()

	s.colliders = data.Colliders
	s.logger.Info("Received static data", "colliders", len(s.colliders))
}

func (s *SinglePlayerState) handleGameInput(input terminal.InputEvent) {
	prevInput := s.currentInput

	switch input {
	case terminal.InputMoveForward:
		s.currentInput.MoveUp = true
		s.currentInput.MoveDown = false
	case terminal.InputMoveBackward:
		s.currentInput.MoveUp = false
		s.currentInput.MoveDown = true
	case terminal.InputMoveLeft:
		s.currentInput.MoveLeft = true
		s.currentInput.MoveRight = false
	case terminal.InputMoveRight:
		s.currentInput.MoveLeft = false
		s.currentInput.MoveRight = true
	case terminal.InputTurnLeft:
		s.currentInput.RotateLeft = true
		s.currentInput.RotateRight = false
	case terminal.InputTurnRight:
		s.currentInput.RotateLeft = false
		s.currentInput.RotateRight = true
	case terminal.InputNone:
		s.currentInput = ports.PlayerInput{}
	}

	if s.currentInput != prevInput {
		s.inputChanged = true
	}
}

func (s *SinglePlayerState) sendInputIfChanged() {
	if !s.inputChanged {
		return
	}

	err := s.client.SendInput(s.currentInput)
	if err != nil {
		s.logger.Error("Failed to send input", "error", err)
	}
	s.inputChanged = false
}

func (s *SinglePlayerState) cleanup() {
	if s.client != nil {
		s.client.Close()
	}
}

func (s *SinglePlayerState) Draw(buf *bytes.Buffer, width, height int) {
	locale := terminal.AppDefaultConfig.Locale

	switch s.phase {
	case PhaseConnecting:
		s.drawCenteredMessage(buf, width, height, locale.SPConnecting)
	case PhaseWaitingForRoom:
		s.drawCenteredMessage(buf, width, height, locale.SPWaitingRoom)
	case PhaseJoiningRoom:
		s.drawCenteredMessage(buf, width, height, locale.SPJoiningRoom)
	case PhaseError:
		s.drawErrorScreen(buf, width, height)
	case PhasePlaying:
		s.drawGameView(buf, width, height)
	}
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

func (s *SinglePlayerState) drawErrorScreen(buf *bytes.Buffer, width, height int) {
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

func (s *SinglePlayerState) drawGameView(buf *bytes.Buffer, width, height int) {
	s.dataMu.RLock()
	playerX := s.playerX
	playerY := s.playerY
	playerDir := s.playerDir
	colliders := s.colliders
	s.dataMu.RUnlock()

	numRays := width
	results := raycast.CastRays(playerX, playerY, playerDir, colliders, numRays)

	s.renderer = raycast.NewRenderer(width, height-2)
	s.renderer.Render(results)

	buf.WriteString(setGreenFont)
	s.renderer.WriteToBuffer(buf)

	locale := terminal.AppDefaultConfig.Locale
	statusLine := fmt.Sprintf("X:%.1f Y:%.1f Dir:%.2f | %s", playerX, playerY, playerDir, locale.SPStatusHint)
	drawCenteredLine(buf, width, statusLine)

	buf.WriteString(resetFontColor)
}
