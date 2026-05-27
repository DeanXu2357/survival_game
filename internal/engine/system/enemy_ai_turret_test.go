package system

import (
	"testing"

	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/system/ai"
	"survival/internal/engine/system/ai/turret"
)

// turretRig wires EnemyAISystem (with a StationaryTurret behavior) in the
// same order as engine.NewGame: AI dispatch before WeaponFireSystem. It
// exists so we can drive a turret tick-by-tick without spinning up the
// full game loop.
type turretRig struct {
	world      *state.World
	tick       *ports.Tick
	enemyAI    *EnemyAISystem
	fireSys    *WeaponFireSystem
	projectile *ProjectileSystem
	revive     *ReviveSystem
	turretID   state.EntityID
	targetID   state.EntityID
	turretDef  state.EntityID
}

func newTurretRig(t *testing.T, targetPos state.Position) *turretRig {
	t.Helper()
	world := state.NewWorld(5.0, 40, 40)
	world.Entity.Alloc() // reserve EntityID(0)

	fistDef := createFistDef(t, world)
	turretDef := createWeaponItemDef(t, world, "Turret", state.TurretSpec)

	// Turret entity at origin, facing +Y (Direction = π).
	turretID, ok := world.CreatePlayer(state.CreatePlayer{
		Position:      state.Position{X: 0, Y: 0},
		Direction:     state.Direction(3.141592653589793),
		MovementSpeed: 0,
		RotationSpeed: 0,
		Radius:        0.6,
		Health:        100,
		FistDefID:     fistDef,
	})
	if !ok {
		t.Fatal("failed to create turret entity")
	}
	world.ApplyCommands()

	inv, _ := world.Inventory.Get(turretID)
	inv.Weapons[2] = state.WeaponSlot{ItemID: turretDef}
	inv.CurrentWeaponIndex = 2
	world.UpdatePlayer(turretID, state.UpdatePlayer{
		UpdateMeta: state.ComponentInventory | state.ComponentRevive,
		Inventory:  inv,
		Revive: state.Revive{
			SpawnHealth:       100,
			RespawnDelayTicks: ports.TicksFromSeconds(1.0),
		},
	})
	world.ApplyCommands()
	world.AI.Upsert(turretID, state.AIBehavior{Key: turret.StationaryTurretKey})

	// Target — circle collider with Health, just like a player.
	targetID := createPlayerForInventory(t, world, targetPos, 0, fistDef)

	tick := new(ports.Tick)
	turretBhv := turret.NewStationaryTurretBehavior()
	registry := map[string]ai.Behavior{turret.StationaryTurretKey: turretBhv}

	rig := &turretRig{
		world:      world,
		tick:       tick,
		enemyAI:    NewEnemyAISystem(world, tick, registry),
		fireSys:    NewWeaponFireSystem(world, tick),
		projectile: NewProjectileSystem(world, tick),
		revive:     NewReviveSystem(world, tick),
		turretID:   turretID,
		targetID:   targetID,
		turretDef:  turretDef,
	}
	return rig
}

func (r *turretRig) step() {
	*r.tick++
	r.world.SyncInputBuffer()
	r.enemyAI.Update(1.0 / 60.0)
	r.fireSys.Update(1.0 / 60.0)
	r.projectile.Update(1.0 / 60.0)
	r.revive.Update(1.0 / 60.0)
	r.world.ApplyCommands()
}

func (r *turretRig) projectileCount() int {
	n := 0
	for range r.world.Projectile.All() {
		n++
	}
	return n
}

func TestTurret_FiresAfterAimDuration(t *testing.T) {
	// First step (tick 1) acquires the target. Aim resolves when
	// (currentTick - aimStart) >= 60, i.e., at tick 61. So 61 steps fire.
	rig := newTurretRig(t, state.Position{X: 0, Y: 5})

	for i := 0; i < 61; i++ {
		rig.step()
	}

	if got := rig.projectileCount(); got != 1 {
		t.Fatalf("expected 1 projectile after 61 ticks, got %d", got)
	}
}

func TestTurret_DoesNotFireBeforeAimComplete(t *testing.T) {
	rig := newTurretRig(t, state.Position{X: 0, Y: 5})

	for i := 0; i < 60; i++ {
		rig.step()
	}

	if got := rig.projectileCount(); got != 0 {
		t.Fatalf("expected 0 projectiles at tick 60, got %d", got)
	}
}

func TestTurret_RespectsCooldownBetweenShots(t *testing.T) {
	rig := newTurretRig(t, state.Position{X: 0, Y: 3})
	rig.world.Health.Set(rig.targetID, 100000)

	// Fire #1 at tick 61, cooldown until tick 91, fire #2 at tick 151.
	for i := 0; i < 152; i++ {
		rig.step()
	}

	if got := rig.projectileCount(); got < 1 {
		t.Fatalf("expected at least 1 projectile, got %d", got)
	}
	inv, _ := rig.world.Inventory.Get(rig.turretID)
	last := inv.Weapons[inv.CurrentWeaponIndex].LastFireTick
	if last < 151 {
		t.Fatalf("expected LastFireTick >= 151 (second shot fired), got %d", last)
	}
}

func TestTurret_NoFireWhenTargetOutsideCone(t *testing.T) {
	rig := newTurretRig(t, state.Position{X: 0, Y: -5})

	for i := 0; i < 200; i++ {
		rig.step()
	}

	if got := rig.projectileCount(); got != 0 {
		t.Fatalf("expected 0 projectiles when target is behind turret, got %d", got)
	}
}

func TestTurret_CancelsAimWhenTargetLeavesCone(t *testing.T) {
	rig := newTurretRig(t, state.Position{X: 0, Y: 5})

	for i := 0; i < 30; i++ {
		rig.step()
	}
	rig.world.Position.Set(rig.targetID, state.Position{X: 0, Y: -5})

	for i := 0; i < 200; i++ {
		rig.step()
	}

	if got := rig.projectileCount(); got != 0 {
		t.Fatalf("expected 0 projectiles when target left cone, got %d", got)
	}
}

func TestTurret_DeadTurretDoesNotFireOrIndicate(t *testing.T) {
	rig := newTurretRig(t, state.Position{X: 0, Y: 5})

	rig.world.Health.Set(rig.turretID, 0)
	for i := 0; i < 60; i++ {
		rig.step()
	}
	if got := rig.projectileCount(); got != 0 {
		t.Fatalf("dead turret should not fire, got %d projectiles", got)
	}
	if len(rig.enemyAI.Indicators()) != 0 {
		t.Fatalf("dead turret should not produce indicators")
	}

	// ReviveSystem respawn delay is 1s; step until past it, then resume.
	for i := 0; i < 70; i++ {
		rig.step()
	}
	if len(rig.enemyAI.Indicators()) != 1 {
		t.Fatalf("expected 1 indicator after revive, got %d", len(rig.enemyAI.Indicators()))
	}
}
