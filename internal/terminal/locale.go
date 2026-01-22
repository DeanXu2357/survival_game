package terminal

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

	SettingScreenSize string
	SettingLanguage   string
	SizeSmall         string
	SizeMedium        string
	SizeLarge         string
	LangNameTW        string
	LangNameEN        string

	SPConnecting   string
	SPWaitingRoom  string
	SPJoiningRoom  string
	SPDisconnected string
	SPError        string
	SPStatusHint   string
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
		SettingsTitle: "設定 (SETTINGS)",
		SettingsHint:  "方向鍵選擇，Enter 變更，Esc 返回",
		MenuTitle:     "SURVIVAL",
		MenuStart:     "開始遊戲",
		MenuMulti:     "多人遊戲",
		MenuSettings:  "設定",
		MenuExit:      "離開",
		MenuHint:      "方向鍵選擇，Enter 確認",

		SettingScreenSize: "螢幕尺寸",
		SettingLanguage:   "語言",
		SizeSmall:         "120 x 32 (小)",
		SizeMedium:        "160 x 40 (中)",
		SizeLarge:         "200 x 50 (大)",
		LangNameTW:        "繁體中文",
		LangNameEN:        "English",

		SPConnecting:   "連線中...",
		SPWaitingRoom:  "等待房間...",
		SPJoiningRoom:  "加入房間中...",
		SPDisconnected: "連線中斷",
		SPError:        "錯誤",
		SPStatusHint:   "WASD 移動, Q/E 轉向, ESC 返回",
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
		SettingsTitle: "SETTINGS",
		SettingsHint:  "Arrows to select, Enter to change, Esc to back",
		MenuTitle:     "SURVIVAL",
		MenuStart:     "Start Game",
		MenuMulti:     "Multiplayer",
		MenuSettings:  "Settings",
		MenuExit:      "Exit",
		MenuHint:      "Arrows to select, Enter to confirm",

		SettingScreenSize: "Screen Size",
		SettingLanguage:   "Language",
		SizeSmall:         "120 x 32 (Small)",
		SizeMedium:        "160 x 40 (Medium)",
		SizeLarge:         "200 x 50 (Large)",
		LangNameTW:        "繁體中文",
		LangNameEN:        "English",

		SPConnecting:   "Connecting...",
		SPWaitingRoom:  "Waiting for room...",
		SPJoiningRoom:  "Joining room...",
		SPDisconnected: "Disconnected",
		SPError:        "Error",
		SPStatusHint:   "WASD move, Q/E turn, ESC back",
	}
)
