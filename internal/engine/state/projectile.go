package state

type ProjectileData struct {
	Speed     float64 // units per second
	Range     float64 // weapon range (used to calculate TTL at spawn)
	Damage    int
	OwnerID   EntityID
	Height    float64 // fixed at 1.5 for now
	ExpiredAt uint64  // game tick at which projectile expires
}
