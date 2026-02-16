package system

import (
	"testing"

	"survival/internal/engine/state"
)

func TestWeaponSwitch_CyclesWeapons(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 100

	playerID := createPlayerWithWeapons(t, world, state.Position{X: 50, Y: 50}, 0)

	ws := NewWeaponSwitchSystem(world, &tick)

	// Default weapon is Gun (index 2)
	weaponState, ok := world.WeaponState.Get(playerID)
	if !ok {
		t.Fatal("WeaponState not found")
	}
	if weaponState.CurrentWeaponIndex != 2 {
		t.Fatalf("Expected default weapon index 2, got %d", weaponState.CurrentWeaponIndex)
	}

	// Switch: 2 -> 0
	world.SetInput(playerID, state.Input{SwitchWeapon: true})
	world.SyncInputBuffer()
	ws.Update(1.0 / 60.0)
	world.ApplyCommands()

	weaponState, _ = world.WeaponState.Get(playerID)
	if weaponState.CurrentWeaponIndex != 0 {
		t.Errorf("Expected weapon index 0 after first switch, got %d", weaponState.CurrentWeaponIndex)
	}

	// Advance past cooldown and switch: 0 -> 1
	tick += weaponSwitchCooldownTicks
	world.SetInput(playerID, state.Input{SwitchWeapon: true})
	world.SyncInputBuffer()
	ws.Update(1.0 / 60.0)
	world.ApplyCommands()

	weaponState, _ = world.WeaponState.Get(playerID)
	if weaponState.CurrentWeaponIndex != 1 {
		t.Errorf("Expected weapon index 1 after second switch, got %d", weaponState.CurrentWeaponIndex)
	}

	// Advance past cooldown and switch: 1 -> 2
	tick += weaponSwitchCooldownTicks
	world.SetInput(playerID, state.Input{SwitchWeapon: true})
	world.SyncInputBuffer()
	ws.Update(1.0 / 60.0)
	world.ApplyCommands()

	weaponState, _ = world.WeaponState.Get(playerID)
	if weaponState.CurrentWeaponIndex != 2 {
		t.Errorf("Expected weapon index 2 after third switch, got %d", weaponState.CurrentWeaponIndex)
	}
}

func TestWeaponSwitch_CooldownEnforced(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 100

	playerID := createPlayerWithWeapons(t, world, state.Position{X: 50, Y: 50}, 0)

	ws := NewWeaponSwitchSystem(world, &tick)

	// First switch: 2 -> 0
	world.SetInput(playerID, state.Input{SwitchWeapon: true})
	world.SyncInputBuffer()
	ws.Update(1.0 / 60.0)
	world.ApplyCommands()

	weaponState, _ := world.WeaponState.Get(playerID)
	if weaponState.CurrentWeaponIndex != 0 {
		t.Fatalf("Expected weapon index 0, got %d", weaponState.CurrentWeaponIndex)
	}

	// Try to switch again immediately (within cooldown) - should be blocked
	tick += 1
	world.SetInput(playerID, state.Input{SwitchWeapon: true})
	world.SyncInputBuffer()
	ws.Update(1.0 / 60.0)
	world.ApplyCommands()

	weaponState, _ = world.WeaponState.Get(playerID)
	if weaponState.CurrentWeaponIndex != 0 {
		t.Errorf("Expected weapon index 0 (cooldown blocked), got %d", weaponState.CurrentWeaponIndex)
	}
}

func TestWeaponSwitch_NoSwitchWithoutInput(t *testing.T) {
	world := setupProjectileWorld()
	var tick uint64 = 100

	playerID := createPlayerWithWeapons(t, world, state.Position{X: 50, Y: 50}, 0)

	ws := NewWeaponSwitchSystem(world, &tick)

	// No SwitchWeapon input
	world.SetInput(playerID, state.Input{Fire: true})
	world.SyncInputBuffer()
	ws.Update(1.0 / 60.0)
	world.ApplyCommands()

	weaponState, _ := world.WeaponState.Get(playerID)
	if weaponState.CurrentWeaponIndex != 2 {
		t.Errorf("Expected weapon index 2 (unchanged), got %d", weaponState.CurrentWeaponIndex)
	}
}
