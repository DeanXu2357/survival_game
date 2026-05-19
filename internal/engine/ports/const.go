package ports

type Tick uint64

func TicksFromSeconds(s float64) Tick {
	return Tick(s * TargetTickRate)
}

func (t Tick) Seconds() float64 {
	return float64(t) / TargetTickRate
}

// FIXME: put const here temporarily, for avoiding circular imports
const (
	TargetTickRate = 60.0
	DeltaTime      = 1.0 / TargetTickRate

	// MaxFrameTime caps the delta time to prevent physics explosions
	// on lag spikes or after pause/resume. Set to 5 frames worth of time.
	MaxFrameTime = 5.0 / TargetTickRate // ~0.0833 seconds (83ms, 5 frames at 60 FPS)

	ItemSlotCount = 6
	PickupRange   = 2.0
)

var (
	NormalReloadTicks = TicksFromSeconds(1.5)
	FastReloadTicks   = TicksFromSeconds(1.0)
)
