package cmd

import (
	"fmt"
	"os"
	"time"
	"unicode"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	rootCmd.AddCommand(termCmd)
}

// termCmd represents the term command
var termCmd = &cobra.Command{
	Use:   "term",
	Short: "A terminal frontend for the survival game",
	Long:  ``,
	Run:   RunTerm,
}

type SizePreset struct {
	Width  int
	Height int
	Label  string
}

var SizePresets = []SizePreset{
	{80, 24, "Small (80x24)"},
	{120, 32, "Medium (120x32)"},
	{160, 40, "Large (160x40)"},
}

var (
	TargetHeight = 32
	TargetWidth  = 120
)

var CurrentLang = LangTW

func RunTerm(cmd *cobra.Command, args []string) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer term.Restore(fd, oldState)

	os.Stdout.WriteString("\033[?25l")
	defer os.Stdout.WriteString("\033[?25h")

	inputChan := make(chan byte)
	go func() {
		b := make([]byte, 1)
		for {
			os.Stdin.Read(b)
			inputChan <- b[0]
		}
	}()

	if !runSettingsScreen(fd, inputChan) {
		return
	}

	if !runResizeScreen(fd, inputChan) {
		return
	}

	for {
		action := runMainMenu(fd, inputChan)
		switch action {
		case MenuActionStart:
			os.Stdout.WriteString("\033[2J\033[H")
			fmt.Println("Starting game...")
			time.Sleep(2 * time.Second)
		case MenuActionMultiplayer:
			os.Stdout.WriteString("\033[2J\033[H")
			fmt.Println("Multiplayer coming soon...")
			time.Sleep(2 * time.Second)
		case MenuActionSettings:
			if !runSettingsScreen(fd, inputChan) {
				return
			}
			if !runResizeScreen(fd, inputChan) {
				return
			}
		case MenuActionExit:
			return
		}
	}
}

func runSettingsScreen(fd int, inputChan chan byte) bool {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	selectedIdx := 1
	boxWidth := 60

	for {
		select {
		case key := <-inputChan:
			if key == 'q' || key == 3 {
				return false
			}

			if key == 13 {
				TargetWidth = SizePresets[selectedIdx].Width
				TargetHeight = SizePresets[selectedIdx].Height
				return true
			}

			if key == 27 {
				nextKey := <-inputChan
				if nextKey == '[' {
					arrowKey := <-inputChan
					if arrowKey == 'A' && selectedIdx > 0 {
						selectedIdx--
					} else if arrowKey == 'B' && selectedIdx < len(SizePresets)-1 {
						selectedIdx++
					}
				}
			}

			if key == 'k' && selectedIdx > 0 {
				selectedIdx--
			} else if key == 'j' && selectedIdx < len(SizePresets)-1 {
				selectedIdx++
			}

			if key == 'l' || key == 'L' {
				if CurrentLang.TitleReady == LangTW.TitleReady {
					CurrentLang = LangEN
				} else {
					CurrentLang = LangTW
				}
			}

		case <-ticker.C:
			width, height, _ := term.GetSize(fd)
			os.Stdout.WriteString("\033[2J\033[H")

			borderLine := ""
			for i := 0; i < boxWidth-2; i++ {
				borderLine += CurrentLang.BoxBorderH
			}
			borderTop := fmt.Sprintf("╔%s╗", borderLine)
			borderBot := fmt.Sprintf("╚%s╝", borderLine)
			emptyRow := DrawBoxRow("", boxWidth)

			fmt.Print("\033[36m")
			fmt.Printf("%s\r\n", borderTop)
			fmt.Printf("%s\r\n", DrawBoxRow(CurrentLang.SettingsTitle, boxWidth))
			fmt.Printf("%s\r\n", emptyRow)
			fmt.Printf("%s\r\n", FormatRow(CurrentLang.LblCurrent, fmt.Sprintf("%d x %d", width, height), boxWidth))
			fmt.Printf("%s\r\n", emptyRow)

			for i, preset := range SizePresets {
				marker := "  "
				if i == selectedIdx {
					marker = "> "
				}
				fmt.Printf("%s\r\n", DrawBoxRow(fmt.Sprintf("%s%s", marker, preset.Label), boxWidth))
			}

			fmt.Printf("%s\r\n", emptyRow)
			fmt.Printf("%s\r\n", DrawBoxRow(CurrentLang.SettingsHint, boxWidth))
			fmt.Printf("%s\r\n", borderBot)
			fmt.Printf("\r\n%s\r\n", PadCenter(CurrentLang.HintToggle, boxWidth))
			fmt.Print("\033[0m")
		}
	}
}

func runResizeScreen(fd int, inputChan chan byte) bool {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	readyToStart := false
	boxWidth := TargetWidth * 3 / 4
	if boxWidth < 60 {
		boxWidth = 60
	}

	for {
		select {
		case key := <-inputChan:
			if key == 'q' || key == 3 {
				return false
			}
			if key == 13 && readyToStart {
				return true
			}

			if key == 'l' || key == 'L' {
				if CurrentLang.TitleReady == LangTW.TitleReady {
					CurrentLang = LangEN
				} else {
					CurrentLang = LangTW
				}
			}

		case <-ticker.C:
			width, height, _ := term.GetSize(fd)
			readyToStart = width >= TargetWidth && height >= TargetHeight

			os.Stdout.WriteString("\033[2J\033[H")

			var title, msg string
			color := "\033[31m"

			if readyToStart {
				color = "\033[32m"
				title = CurrentLang.TitleReady
				msg = CurrentLang.MsgReady
			} else {
				title = CurrentLang.TitleWait
				msg = CurrentLang.MsgWait
			}

			borderLine := ""
			for i := 0; i < boxWidth-2; i++ {
				borderLine += CurrentLang.BoxBorderH
			}
			borderTop := fmt.Sprintf("╔%s╗", borderLine)
			borderBot := fmt.Sprintf("╚%s╝", borderLine)
			emptyRow := DrawBoxRow("", boxWidth)

			fmt.Print(color)
			fmt.Printf("%s\r\n", borderTop)
			fmt.Printf("%s\r\n", DrawBoxRow(title, boxWidth))
			fmt.Printf("%s\r\n", emptyRow)
			fmt.Printf("%s\r\n", FormatRow(CurrentLang.LblCurrent, fmt.Sprintf("%d x %d", width, height), boxWidth))
			fmt.Printf("%s\r\n", FormatRow(CurrentLang.LblTarget, fmt.Sprintf("%d x %d", TargetWidth, TargetHeight), boxWidth))
			fmt.Printf("%s\r\n", emptyRow)
			fmt.Printf("%s\r\n", DrawBoxRow(msg, boxWidth))
			fmt.Printf("%s\r\n", borderBot)
			fmt.Printf("\r\n%s\r\n", PadCenter(CurrentLang.HintToggle, boxWidth))
			fmt.Print("\033[0m")

			if readyToStart {
				fmt.Printf("\033[%d;%dH\033[32m[OK]\033[0m", height, width-3)
			}
		}
	}
}

type MenuAction int

const (
	MenuActionNone MenuAction = iota
	MenuActionStart
	MenuActionMultiplayer
	MenuActionSettings
	MenuActionExit
)

func runMainMenu(fd int, inputChan chan byte) MenuAction {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	selectedIdx := 0
	boxWidth := TargetWidth
	boxHeight := TargetHeight - 2

	menuItems := []struct {
		Label  func() string
		Action MenuAction
	}{
		{func() string { return CurrentLang.MenuStart }, MenuActionStart},
		{func() string { return CurrentLang.MenuMulti }, MenuActionMultiplayer},
		{func() string { return CurrentLang.MenuSettings }, MenuActionSettings},
		{func() string { return CurrentLang.MenuExit }, MenuActionExit},
	}

	for {
		select {
		case key := <-inputChan:
			if key == 'q' || key == 3 {
				return MenuActionExit
			}

			if key == 13 {
				return menuItems[selectedIdx].Action
			}

			if key == 27 {
				nextKey := <-inputChan
				if nextKey == '[' {
					arrowKey := <-inputChan
					if arrowKey == 'A' && selectedIdx > 0 {
						selectedIdx--
					} else if arrowKey == 'B' && selectedIdx < len(menuItems)-1 {
						selectedIdx++
					}
				}
			}

			if key == 'k' && selectedIdx > 0 {
				selectedIdx--
			} else if key == 'j' && selectedIdx < len(menuItems)-1 {
				selectedIdx++
			}

			if key == 'l' || key == 'L' {
				if CurrentLang.TitleReady == LangTW.TitleReady {
					CurrentLang = LangEN
				} else {
					CurrentLang = LangTW
				}
			}

		case <-ticker.C:
			os.Stdout.WriteString("\033[2J\033[H")

			borderLine := ""
			for i := 0; i < boxWidth-2; i++ {
				borderLine += CurrentLang.BoxBorderH
			}
			borderTop := fmt.Sprintf("╔%s╗", borderLine)
			borderBot := fmt.Sprintf("╚%s╝", borderLine)
			emptyRow := DrawBoxRow("", boxWidth)

			fixedRows := 2 + 1 + 1 + len(menuItems) + 1 + 1
			remainingRows := boxHeight - fixedRows
			if remainingRows < 0 {
				remainingRows = 0
			}
			topPadding := remainingRows / 2
			bottomPadding := remainingRows - topPadding

			fmt.Print("\033[33m")
			fmt.Printf("%s\r\n", borderTop)

			for i := 0; i < topPadding; i++ {
				fmt.Printf("%s\r\n", emptyRow)
			}

			fmt.Printf("%s\r\n", DrawBoxRow(CurrentLang.MenuTitle, boxWidth))
			fmt.Printf("%s\r\n", emptyRow)

			for i, item := range menuItems {
				marker := "  "
				if i == selectedIdx {
					marker = "> "
				}
				fmt.Printf("%s\r\n", DrawBoxRow(fmt.Sprintf("%s%s", marker, item.Label()), boxWidth))
			}

			fmt.Printf("%s\r\n", emptyRow)
			fmt.Printf("%s\r\n", DrawBoxRow(CurrentLang.MenuHint, boxWidth))

			for i := 0; i < bottomPadding; i++ {
				fmt.Printf("%s\r\n", emptyRow)
			}

			fmt.Printf("%s\r\n", borderBot)
			fmt.Print("\033[0m")
		}
	}
}

type LocaleData struct {
	TitleReady    string
	TitleWait     string
	LblCurrent    string
	LblTarget     string
	MsgReady      string
	MsgWait       string
	HintToggle    string
	BoxBorderH    string
	BoxBorderV    string
	SettingsTitle string
	SettingsHint  string
	MenuTitle     string
	MenuStart     string
	MenuMulti     string
	MenuSettings  string
	MenuExit      string
	MenuHint      string
}

var (
	LangTW = LocaleData{
		TitleReady:    "系統檢測通過 (SYSTEM READY)",
		TitleWait:     "視窗尺寸不足 (WINDOW TOO SMALL)",
		LblCurrent:    "目前尺寸",
		LblTarget:     "目標尺寸",
		MsgReady:      "[ 請按 Enter 鍵繼續 ]",
		MsgWait:       "--> 請拉大視窗，直到此訊息變為綠色 <--",
		HintToggle:    "(按 'L' 切換語言 / Toggle Language)",
		BoxBorderH:    "═",
		BoxBorderV:    "║",
		SettingsTitle: "選擇目標尺寸 (SELECT SIZE)",
		SettingsHint:  "方向鍵選擇，Enter 確認",
		MenuTitle:     "SURVIVAL",
		MenuStart:     "Start Game",
		MenuMulti:     "Multiplayer",
		MenuSettings:  "Settings",
		MenuExit:      "Exit",
		MenuHint:      "方向鍵選擇，Enter 確認",
	}

	LangEN = LocaleData{
		TitleReady:    "SYSTEM READY",
		TitleWait:     "WINDOW TOO SMALL",
		LblCurrent:    "Current Size",
		LblTarget:     "Target Size ",
		MsgReady:      "[ Press Enter to Continue ]",
		MsgWait:       "--> Please enlarge the window <--",
		HintToggle:    "(Press 'L' to Toggle Language)",
		BoxBorderH:    "=",
		BoxBorderV:    "|",
		SettingsTitle: "SELECT SIZE",
		SettingsHint:  "Arrows to select, Enter to confirm",
		MenuTitle:     "SURVIVAL",
		MenuStart:     "Start Game",
		MenuMulti:     "Multiplayer",
		MenuSettings:  "Settings",
		MenuExit:      "Exit",
		MenuHint:      "Arrows to select, Enter to confirm",
	}
)

func CalcVisualWidth(s string) int {
	width := 0
	for _, r := range s {
		if unicode.Is(unicode.Han, r) || (r > 127 && r != '═' && r != '║') {
			width += 2
		} else {
			width += 1
		}
	}
	return width
}

func PadCenter(text string, totalWidth int) string {
	visLen := CalcVisualWidth(text)
	padding := totalWidth - visLen
	if padding < 0 {
		return text
	}
	leftPad := padding / 2
	rightPad := padding - leftPad

	return fmt.Sprintf("%*s%s%*s", leftPad, "", text, rightPad, "")
}

func FormatRow(label string, value string, boxWidth int) string {
	contentWidth := boxWidth - 4

	fullStr := fmt.Sprintf("%s: %s", label, value)

	visLen := CalcVisualWidth(fullStr)
	padLen := contentWidth - visLen

	if padLen < 0 {
		padLen = 0
	}

	return fmt.Sprintf("%s %s%*s %s",
		CurrentLang.BoxBorderV,
		fullStr,
		padLen, "",
		CurrentLang.BoxBorderV)
}

func DrawBoxRow(text string, boxWidth int) string {
	contentWidth := boxWidth - 4
	centeredText := PadCenter(text, contentWidth)
	return fmt.Sprintf("%s %s %s", CurrentLang.BoxBorderV, centeredText, CurrentLang.BoxBorderV)
}
