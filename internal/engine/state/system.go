package state

type System interface {
	Update(dt float64)
	ReadMeta() Meta
	WriteMeta() Meta
}

type SystemManager struct {
	world   *World
	systems []System
}

func NewSystemManager(world *World) *SystemManager {
	return &SystemManager{
		world:   world,
		systems: make([]System, 0),
	}
}

func (sm *SystemManager) Register(system System) {
	sm.systems = append(sm.systems, system)
}

func (sm *SystemManager) Update(dt float64) {
	for _, sys := range sm.systems {
		sys.Update(dt)
	}
}

func (sm *SystemManager) World() *World {
	return sm.world
}
