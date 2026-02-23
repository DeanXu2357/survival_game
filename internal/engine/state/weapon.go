package state

import "survival/internal/engine/ports"

type WeaponType uint8

const (
	WeaponTypeFist WeaponType = iota
	WeaponTypeKnife
	WeaponTypeGun
)

// AmmoCategory links weapons with compatible magazines.
type AmmoCategory uint8

const (
	AmmoCategoryNone   AmmoCategory = iota // melee weapons
	AmmoCategoryPistol                     // pistol magazines
)

// WeaponConfig defines the immutable stats for a weapon type.
// Future: refactor to Flyweight pattern to share specs via pointers.
type WeaponConfig struct {
	Type         WeaponType
	Range        float64 // max range in world units
	FireRate     float64 // shots per second
	Damage       int
	Speed        float64      // projectile speed in units per second
	AmmoCategory AmmoCategory // AmmoCategoryNone for melee
}

// FistSpec defines the default fist weapon stats.
var FistSpec = WeaponConfig{Type: WeaponTypeFist, Range: 5, FireRate: 3, Damage: 5, Speed: 50}

// ReloadType indicates which reload action is in progress.
type ReloadType uint8

const (
	ReloadTypeNone   ReloadType = iota
	ReloadTypeNormal            // 1.5s (90 ticks at 60 FPS)
	ReloadTypeFast              // 1.0s (60 ticks at 60 FPS)
)

// WeaponSlot represents a weapon equipped in the player's weapon loadout.
type WeaponSlot struct {
	ItemID          EntityID   // 0 = empty (slot 0 always has FistDefID)
	LastFireTick    ports.Tick // per-weapon fire cooldown
	LoadedMagID     EntityID   // which magazine type is loaded (0 = none)
	LoadedMagAmmo   int        // bullets remaining in loaded magazine
	MagCapacity     int        // max bullets for loaded magazine (copied from ItemConfig at load time)
	ReloadStartTick ports.Tick // tick when reload began (0 = not reloading)
	ReloadType      ReloadType // which reload type is in progress
}

func (s WeaponSlot) IsEmpty() bool { return s.ItemID == 0 }
