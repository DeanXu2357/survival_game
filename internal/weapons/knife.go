package weapons

type Knife struct {
	ID    string
	Range float64
}

func (k *Knife) GetID() string       { return k.ID }
func (k *Knife) GetType() WeaponType { return WeaponTypeKnife }
func (k *Knife) CanUse() bool        { return true }
