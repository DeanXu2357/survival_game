package state

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"survival/internal/adapters/repository/maploader"
	"survival/internal/engine"
	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/terminal"
	"survival/internal/terminal/session"
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
	menuItemCount := 5

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
		case 0:
			return s.startSinglePlayer()
		case 1:
			return s.startPracticeRange()
		case 2:
			return terminal.Command{Type: terminal.CmdPush, NextState: NewMultiplayerState(s.fd, s.logger)}
		case 3:
			return terminal.Command{Type: terminal.CmdPush, NextState: NewSettingState(s.fd, s.logger)}
		case 4:
			return terminal.Command{Type: terminal.CmdQuit}
		}

	case terminal.InputCancel:
		return terminal.Command{Type: terminal.CmdQuit}
	}

	return terminal.Command{Type: terminal.CmdNone}
}

func (s *MainMenuState) startSinglePlayer() terminal.Command {
	return s.startGame(false)
}

func (s *MainMenuState) startPracticeRange() terminal.Command {
	return s.startGame(true)
}

func (s *MainMenuState) startGame(spawnDummy bool) terminal.Command {
	mapConfig := loadMapOrDefault(s.logger)

	game, err := engine.NewGame(mapConfig)
	if err != nil {
		s.logger.Error("Failed to create game", "error", err)
		return terminal.Command{Type: terminal.CmdNone}
	}

	entityID, err := game.JoinPlayer()
	if err != nil {
		s.logger.Error("Failed to join player", "error", err)
		return terminal.Command{Type: terminal.CmdNone}
	}

	knifeDefID, err := game.RegisterItemDef(state.ItemConfig{
		Name:     "Knife",
		Type:     state.ItemTypeWeapon,
		MaxStack: 1,
		WeaponConfig: state.WeaponConfig{
			Type:     state.WeaponTypeKnife,
			Range:    3,
			FireRate: 2,
			Damage:   30,
			Speed:    0,
		},
	})
	if err != nil {
		s.logger.Error("Failed to register knife", "error", err)
		return terminal.Command{Type: terminal.CmdNone}
	}

	gunDefID, err := game.RegisterItemDef(state.ItemConfig{
		Name:     "Pistol",
		Type:     state.ItemTypeWeapon,
		MaxStack: 1,
		WeaponConfig: state.WeaponConfig{
			Type:         state.WeaponTypeGun,
			Range:        50,
			FireRate:     2,
			Damage:       25,
			Speed:        100,
			AmmoCategory: state.AmmoCategoryPistol,
		},
		AmmoCategory: state.AmmoCategoryPistol,
	})
	if err != nil {
		s.logger.Error("Failed to register pistol", "error", err)
		return terminal.Command{Type: terminal.CmdNone}
	}

	magDefID, err := game.RegisterItemDef(state.ItemConfig{
		Name:         "PistolMag",
		Type:         state.ItemTypeMagazine,
		MaxStack:     1,
		MagCapacity:  12,
		AmmoCategory: state.AmmoCategoryPistol,
	})
	if err != nil {
		s.logger.Error("Failed to register pistol magazine", "error", err)
		return terminal.Command{Type: terminal.CmdNone}
	}

	inv, ok := game.PlayerInventory(entityID)
	if !ok {
		s.logger.Error("Failed to get player inventory")
		return terminal.Command{Type: terminal.CmdNone}
	}
	inv.Weapons[1] = state.WeaponSlot{ItemID: knifeDefID}
	inv.Weapons[2] = state.WeaponSlot{
		ItemID:        gunDefID,
		LoadedMagID:   magDefID,
		LoadedMagAmmo: 12,
		MagCapacity:   12,
	}
	inv.Items[0] = state.ItemSlot{ItemDefID: magDefID, Quantity: 1, Ammo: 12}
	inv.Items[1] = state.ItemSlot{ItemDefID: magDefID, Quantity: 1, Ammo: 12}
	if err := game.SetPlayerInventory(entityID, inv); err != nil {
		s.logger.Error("Failed to set player inventory", "error", err)
		return terminal.Command{Type: terminal.CmdNone}
	}

	if spawnDummy {
		if _, err := game.SpawnTrainingDummyNearSpawn(); err != nil {
			s.logger.Warn("Failed to spawn training dummy", "error", err)
		}
	}

	sess := session.NewGameSession(game, entityID)
	colliders := staticEntitiesToColliders(game.WallEntities())

	return terminal.Command{
		Type:      terminal.CmdPush,
		NextState: NewSinglePlayerState(s.fd, s.logger, sess, colliders),
	}
}

func loadMapOrDefault(logger *slog.Logger) *engine.MapConfig {
	loader := maploader.NewJSONMapLoader("./maps")
	mapConfig, err := loader.LoadMap("office_floor_01")
	if err != nil {
		logger.Warn("Failed to load map, using default", "error", err)
		return engine.DefaultMapConfig()
	}
	return mapConfig
}

func staticEntitiesToColliders(statics []state.StaticEntity) []ports.Collider {
	colliders := make([]ports.Collider, len(statics))
	for i, entity := range statics {
		colliders[i] = ports.Collider{
			ID:            uint64(entity.ID),
			X:             entity.Collider.Center.X,
			Y:             entity.Collider.Center.Y,
			HalfX:         entity.Collider.HalfSize.X,
			HalfY:         entity.Collider.HalfSize.Y,
			Radius:        entity.Collider.Radius,
			ShapeType:     uint8(entity.Collider.ShapeType),
			Rotation:      0,
			Height:        entity.Collider.Height,
			BaseElevation: entity.Collider.BaseElevation,
		}
	}
	return colliders
}

func (s *MainMenuState) Draw(buf *bytes.Buffer, width, height int) {
	locale := terminal.AppDefaultConfig.Locale
	boxWidth := 50

	menuItems := []string{
		locale.MenuStart,
		locale.MenuPracticeRange,
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
