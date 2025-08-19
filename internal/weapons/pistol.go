package weapons

import "time"

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
