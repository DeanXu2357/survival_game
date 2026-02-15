package system

import (
	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/vector"
)

const (
	defaultProjectileSpeed  = 20.0
	defaultProjectileRange  = 50.0
	defaultProjectileDamage = 25
	defaultProjectileHeight = 1.5
)

var _ state.System = (*WeaponFireSystem)(nil)

type WeaponFireSystem struct {
	world       *state.World
	currentTick *uint64
}

func NewWeaponFireSystem(world *state.World, currentTick *uint64) *WeaponFireSystem {
	return &WeaponFireSystem{world: world, currentTick: currentTick}
}

func (wf *WeaponFireSystem) ReadMeta() state.Meta {
	return state.ComponentInput | state.ComponentPosition | state.ComponentDirection
}

func (wf *WeaponFireSystem) WriteMeta() state.Meta {
	return state.ComponentProjectile | state.ComponentPosition | state.ComponentDirection | state.ComponentMeta
}

func (wf *WeaponFireSystem) Update(dt float64) {
	world := wf.world

	for entityID, input := range world.Input.All() {
		if !input.Fire {
			// TODO: lack of fire rate control, will cause projectile spam if player holds down fire button
			// must implement fire rate control in the future
			continue
		}

		pos, posOk := world.Position.Get(entityID)
		dir, dirOk := world.Direction.Get(entityID)
		if !posOk || !dirOk {
			continue
		}

		spawnPos := state.Position(vector.Vector2D(pos).Add(vector.Forward(float64(dir)).Scale(0.5)))

		expiredAt := *wf.currentTick + uint64(defaultProjectileRange/defaultProjectileSpeed*ports.TargetTickRate)

		world.CreateProjectileEntity(state.CreateProjectile{
			Position:  spawnPos,
			Direction: dir,
			ProjectileData: state.ProjectileData{
				Speed:     defaultProjectileSpeed,
				Range:     defaultProjectileRange,
				Damage:    defaultProjectileDamage,
				OwnerID:   entityID,
				Height:    defaultProjectileHeight,
				ExpiredAt: expiredAt,
			},
		})
	}
}
