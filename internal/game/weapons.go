package game

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

type Knife struct {
	ID    string
	Range float64
}

func (k *Knife) GetID() string       { return k.ID }
func (k *Knife) GetType() WeaponType { return WeaponTypeKnife }
func (k *Knife) CanUse() bool        { return true }

type Pistol struct {
	ID              string
	CurrentMagazine *Magazine
	Range           float64
	IsReloading     bool
	ReloadStartTime time.Time
	ReloadType      ReloadType
}

func (p *Pistol) GetID() string       { return p.ID }
func (p *Pistol) GetType() WeaponType { return WeaponTypePistol }
func (p *Pistol) CanUse() bool {
	return !p.IsReloading && p.CurrentMagazine != nil && p.CurrentMagazine.CurrentAmmo > 0
}
func (p *Pistol) GetAmmoCount() int {
	if p.CurrentMagazine == nil {
		return 0
	}
	return p.CurrentMagazine.CurrentAmmo
}
func (p *Pistol) GetRange() float64 { return p.Range }

func (p *Pistol) Reload(reloadType ReloadType, availableMagazines []Magazine) bool {
	if p.IsReloading || len(availableMagazines) <= 1 {
		return false
	}

	p.IsReloading = true
	p.ReloadType = reloadType
	p.ReloadStartTime = time.Now()
	return true
}

type Inventory struct {
	MeleeWeapons  []Knife
	RangedWeapons []Pistol
	Magazines     []Magazine
	MaxSlots      int
}

type Projectile struct {
	ID        string
	Position  Vector2D
	Direction Vector2D
	Speed     float64
	Range     float64
	Damage    int
	OwnerID   string
}

func (p *Projectile) GetPosition() Vector2D {
	return p.Position
}
