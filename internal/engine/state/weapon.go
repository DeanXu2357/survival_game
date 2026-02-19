package state

type WeaponType uint8

const (
	WeaponTypeFist WeaponType = iota
	WeaponTypeKnife
	WeaponTypeGun
)

// WeaponConfig defines the immutable stats for a weapon type.
// Future: refactor to Flyweight pattern to share specs via pointers.
type WeaponConfig struct {
	Type     WeaponType
	Range    float64 // max range in world units
	FireRate float64 // shots per second
	Damage   int
	Speed    float64 // projectile speed in units per second
}

// FistSpec defines the default fist weapon stats.
var FistSpec = WeaponConfig{Type: WeaponTypeFist, Range: 5, FireRate: 3, Damage: 5, Speed: 50}

// WeaponSlot represents a weapon equipped in the player's weapon loadout.
type WeaponSlot struct {
	ItemDefID    EntityID // 0 = empty (slot 0 always has FistDefID)
	LastFireTick uint64   // per-weapon fire cooldown
}

func (s WeaponSlot) IsEmpty() bool { return s.ItemDefID == 0 }
