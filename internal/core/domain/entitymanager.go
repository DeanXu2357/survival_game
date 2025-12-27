package domain

import "sync"

const (
	maxEntityCount = 10000 // the max number of entities allowed in the world in one tic
)

type EntityID uint64

func NewEntityID(index int, version uint32) EntityID {
	return EntityID(uint64(version)<<32 | uint64(index))
}

func (id EntityID) Index() int { return int(id & 0xFFFFFFFF) }

func (id EntityID) Version() uint32 { return uint32(id >> 32) }

/*
 * Because of rwLock usage. 'IsAlive' will be called frequently in the future,
 * it could be a performance bottleneck.
 * We can optimize it later if needed by using atomic operations or other lock-free techniques.
 */

type EntityManager struct {
	versions []uint32
	freeList []int
	count    int
	rwLock   sync.RWMutex // not sure if needed
}

func NewEntityManager() *EntityManager {
	return &EntityManager{
		versions: make([]uint32, 0),
		freeList: make([]int, 0),
		count:    0,
	}
}

func (em *EntityManager) Alloc() (EntityID, bool) {
	em.rwLock.Lock()
	defer em.rwLock.Unlock()

	var idx int
	if len(em.freeList) > 0 {
		last := len(em.freeList) - 1
		idx = em.freeList[last]
		em.freeList = em.freeList[:last]
	} else {
		if em.count >= maxEntityCount {
			return 0, false
		}

		idx = len(em.versions)
		em.versions = append(em.versions, 0)
	}

	em.count++
	return NewEntityID(idx, em.versions[idx]), true
}

func (em *EntityManager) Free(e EntityID) bool {
	em.rwLock.Lock()
	defer em.rwLock.Unlock()

	index := e.Index()

	if !em.IsAlive(e) {
		return false
	}

	em.versions[index]++
	em.freeList = append(em.freeList, index)
	em.count--
	return true
}

func (em *EntityManager) IsAlive(e EntityID) bool {
	em.rwLock.RLock()
	defer em.rwLock.RUnlock()

	index := e.Index()
	version := e.Version()
	return index >= 0 && index < len(em.versions) && em.versions[index] == version
}
