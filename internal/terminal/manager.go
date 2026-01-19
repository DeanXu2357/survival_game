package terminal

import (
	"bytes"
	"log/slog"
	"os"
	"time"
)

var AppDefaultConfig = GameConfig{
	Height: 32,
	Width:  120,
	Locale: LangTW,
}

func NewGameManager(fd int, logger *slog.Logger, inputChan chan KeyEvent, initialState GameState) *GameManager {
	gm := &GameManager{
		fd:        fd,
		logger:    logger,
		inputChan: inputChan,
		stack:     []GameState{},
		isRunning: true,
	}

	initialState.Init()
	gm.stack = append(gm.stack, initialState)

	return gm
}

type GameManager struct {
	fd        int
	logger    *slog.Logger
	inputChan chan KeyEvent
	stack     []GameState
	isRunning bool
}

func (gm *GameManager) Run() {
	ticker := time.NewTicker(1 * time.Second / 60)
	defer ticker.Stop()

	for gm.isRunning {
		select {
		case keyEvent := <-gm.inputChan:
			// handle input event (maybe do buffer input to avoid input lag?)
			gm.handleUpdate(parseInput(keyEvent), 0)
		case _ = <-ticker.C:
			// maybe can do drain input channel to optimize input experience
			gm.handleUpdate(InputNone, 0)

			gm.render()
		}
	}
}

func parseInput(input KeyEvent) InputEvent {
	s := string(input)
	switch s {
	case "\033[A": // Up Arrow
		return InputMoveBackward
	case "\033[B": // Down Arrow
		return InputMoveForward
	case "\r", "\n": // Enter
		return InputAction
	case "q", "\033": // q or Esc
		return InputCancel
	}
	return InputNone
}

func (gm *GameManager) handleUpdate(input InputEvent, dt time.Duration) {
	if len(gm.stack) == 0 {
		return
	}

	currentState := gm.stack[len(gm.stack)-1]

	cmd := currentState.Update(input, dt)

	switch cmd.Type {
	case CmdNone:
		// do nothing
	case CmdPush:
		if cmd.NextState != nil {
			cmd.NextState.Init()
			gm.stack = append(gm.stack, cmd.NextState)
		}
	case CmdPop:
		if len(gm.stack) > 1 {
			gm.stack = gm.stack[:len(gm.stack)-1]
		}
	case CmdSwap:
		if len(gm.stack) > 0 && cmd.NextState != nil {
			gm.stack[len(gm.stack)-1] = cmd.NextState
			cmd.NextState.Init()
		}
	case CmdQuit:
		gm.isRunning = false
	default:
		panic("unhandled default case")
	}
}

func (gm *GameManager) render() {
	if len(gm.stack) == 0 {
		return
	}

	currentState := gm.stack[len(gm.stack)-1]

	width := AppDefaultConfig.Width
	height := AppDefaultConfig.Height

	buf := new(bytes.Buffer)

	buf.WriteString(ResetCursor)

	currentState.Draw(buf, width, height)

	buf.WriteString(ClearToEnd)

	os.Stdout.Write(buf.Bytes())
}

type CommandType uint8

const (
	CmdNone CommandType = iota
	CmdPush
	CmdPop
	CmdSwap
	CmdQuit
)

type Command struct {
	Type      CommandType
	NextState GameState
}

type GameConfig struct {
	Height int
	Width  int
	Locale LocaleData
}
