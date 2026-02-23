package system

import (
	"survival/internal/engine/ports"
	"survival/internal/engine/state"
	"survival/internal/engine/vector"
)

var _ state.System = (*ProjectileSystem)(nil)

type ProjectileSystem struct {
	world       *state.World
	currentTick *ports.Tick
}

func NewProjectileSystem(world *state.World, currentTick *ports.Tick) *ProjectileSystem {
	return &ProjectileSystem{world: world, currentTick: currentTick}
}

func (ps *ProjectileSystem) ReadMeta() state.Meta {
	return state.ComponentProjectile | state.ComponentPosition | state.ComponentDirection | state.ComponentPlayerHitbox
}

func (ps *ProjectileSystem) WriteMeta() state.Meta {
	return state.ComponentPosition | state.ComponentHealth
}

func (ps *ProjectileSystem) Update(dt float64) {
	world := ps.world
	tick := *ps.currentTick

	var toDestroy []state.EntityID

	for entityID, proj := range world.Projectile.All() {
		// TTL check
		if tick >= proj.ExpiredAt {
			toDestroy = append(toDestroy, entityID)
			continue
		}

		pos, posOk := world.Position.Get(entityID)
		dir, dirOk := world.Direction.Get(entityID)
		if !posOk || !dirOk {
			toDestroy = append(toDestroy, entityID)
			continue
		}

		// Calculate movement
		velocity := vector.Forward(float64(dir)).Scale(proj.Speed * dt)
		newPos := vector.Vector2D(pos).Add(velocity)

		// Wall collision: point-in-AABB test
		if ps.checkWallCollision(newPos) {
			toDestroy = append(toDestroy, entityID)
			continue
		}

		// Player hit detection
		if ps.checkPlayerHit(entityID, proj, newPos) {
			toDestroy = append(toDestroy, entityID)
			continue
		}

		// Update position via command buffer
		world.UpdatePlayer(entityID, state.UpdatePlayer{
			UpdateMeta: state.ComponentPosition,
			Position:   state.Position(newPos),
		})
	}

	for _, id := range toDestroy {
		world.QueueDestroyEntity(id)
	}
}

// checkWallCollision tests if a point is inside any static wall AABB.
func (ps *ProjectileSystem) checkWallCollision(point vector.Vector2D) bool {
	world := ps.world

	bounds := state.Bounds{
		MinX: point.X, MinY: point.Y,
		MaxX: point.X, MaxY: point.Y,
	}

	for _, cell := range world.Grid.CellsInBounds(bounds) {
		for _, entry := range cell.Entries {
			if !entry.Layer.Has(state.LayerStatic) {
				continue
			}

			collider, ok := world.Collider.Get(entry.EntityID)
			if !ok {
				continue
			}

			wallMin, wallMax := collider.BoundingBox()
			if point.X >= wallMin.X && point.X <= wallMax.X &&
				point.Y >= wallMin.Y && point.Y <= wallMax.Y {
				return true
			}
		}
	}

	return false
}

// checkPlayerHit tests if a projectile point hits any player (except the owner).
// Uses point-in-circle test against PlayerHitbox.
// Applies damage immediately on hit.
func (ps *ProjectileSystem) checkPlayerHit(projectileID state.EntityID, proj state.ProjectileData, point vector.Vector2D) bool {
	world := ps.world

	for entityID, hitbox := range world.PlayerHitbox.All() {
		// Skip owner
		if entityID == proj.OwnerID {
			continue
		}

		// Point-in-circle test
		center := vector.Vector2D(hitbox.Center)
		dist := point.DistanceTo(center)
		if dist <= hitbox.Radius {
			// Apply damage immediately
			health, hOk := world.Health.Get(entityID)
			if hOk {
				newHealth := state.Health(int(health) - proj.Damage)
				world.Health.Set(entityID, newHealth)
			}
			return true
		}
	}

	return false
}
