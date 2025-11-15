package weapons

import "time"

type WeaponType int

const (
	WeaponTypeKnife WeaponType = iota
	WeaponTypePistol
)

type ReloadType int

const (
	NoReload ReloadType = iota
	NormalReload
	FastReload
)

const (
	NormalReloadDuration = 3 * time.Second
	FastReloadDuration   = 1 * time.Second
)

type Weapon interface {
	GetID() string
	GetType() WeaponType
	CanUse() bool
}

type RangedWeapon interface {
	Weapon
	Reload(reloadType ReloadType, availableMagazines []Magazine) bool
	GetAmmoCount() int
	GetRange() float64
}

type Magazine struct {
	ID          string
	CurrentAmmo int
	MaxCapacity int
	IsEmpty     bool
}

type Inventory struct {
	MeleeWeapons  []Knife
	RangedWeapons []Pistol
	Magazines     []Magazine
	MaxSlots      int
}
