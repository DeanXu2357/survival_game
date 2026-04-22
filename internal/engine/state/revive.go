package state

import "survival/internal/engine/ports"

// Revive marks an entity as respawning after death. When Health drops to
// zero the ReviveSystem stamps DeadAt; after RespawnDelayTicks have elapsed
// it refills Health to SpawnHealth and clears DeadAt. The hit-detection
// path skips entities whose Health <= 0, so no component surgery is needed.
type Revive struct {
	SpawnHealth       Health
	RespawnDelayTicks ports.Tick
	DeadAt            ports.Tick // 0 means alive
}
