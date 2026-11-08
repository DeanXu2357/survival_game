package turret

import "survival/internal/engine/system/ai/goap"

const (
	symHasTarget goap.Symbol = 1 << iota
	symAimed
	symWeaponReady
	symTargetDown
)
