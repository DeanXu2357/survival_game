package ports

// FIXME: put const here temporarily, for avoiding circular imports
const (
	TargetTickRate = 60.0
	DeltaTime      = 1.0 / TargetTickRate

	// MaxFrameTime caps the delta time to prevent physics explosions
	// on lag spikes or after pause/resume. Set to 5 frames worth of time.
	MaxFrameTime = 5.0 / TargetTickRate // ~0.0833 seconds (83ms, 5 frames at 60 FPS)

	ItemSlotCount = 6
	PickupRange   = 2.0

	NormalReloadTicks = 90 // 1.5s at 60 FPS
	FastReloadTicks   = 60 // 1.0s at 60 FPS
)
