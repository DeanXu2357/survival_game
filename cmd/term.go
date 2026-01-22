package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"survival/internal/terminal"
	"survival/internal/terminal/state"
)

func init() {
	rootCmd.AddCommand(termCmd)
}

// termCmd represents the term command
var termCmd = &cobra.Command{
	Use:   "term",
	Short: "A terminal frontend for the survival game",
	Long:  ``,
	Run:   RunTermKai,
}

func RunTermKai(cmd *cobra.Command, args []string) {
	logger, logFile, err := newLogger("term.log")
	if err != nil {
		panic(err)
	}
	defer func() {
		if cerr := logFile.Close(); cerr != nil {
			panic(cerr)
		}
	}()

	// set raw mode
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		logger.Error("failed to set terminal to raw mode", "error", err)
		panic(err)
	}
	defer func() {
		if rerr := term.Restore(fd, oldState); rerr != nil {
			logger.Error("failed to restore terminal state", "error", rerr)
			panic(rerr)
		}
	}()

	// hide cursor
	if _, err := os.Stdout.WriteString(terminal.HideCursor); err != nil {
		logger.Error("failed to hide cursor", "error", err)
		panic(err)
	}
	defer func() {
		if _, err := os.Stdout.WriteString(terminal.ShowCursor); err != nil {
			logger.Error("failed to show cursor", "error", err)
			panic(err)
		}
	}()

	os.Stdout.WriteString(terminal.ClearScreen + terminal.ResetCursor)

	inputChan := make(chan terminal.KeyEvent, 500)
	go func() {
		b := make([]byte, 128)
		for {
			n, readErr := os.Stdin.Read(b)
			if readErr != nil {
				logger.Error("failed to read from stdin", "error", readErr)
				return
			}
			if n > 0 {
				event := make([]byte, n)
				copy(event, b[:n])
				inputChan <- event
			}
		}
	}()

	gameManager := terminal.NewGameManager(
		fd,
		logger,
		inputChan,
		state.NewResizeState(fd, logger),
	)

	gameManager.Run()
}

func newLogger(filePath string) (logger *slog.Logger, file *os.File, err error) {
	file, err = os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}

	logger = slog.New(slog.NewJSONHandler(file, nil))
	slog.SetDefault(logger)

	logger.Info("logger initialized", "file", filePath)

	return logger, file, nil
}
