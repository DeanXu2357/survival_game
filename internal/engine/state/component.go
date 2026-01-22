package state

import "iter"

type ComponentManager[T any] struct {
	data            []T
	IndexToEntityID []EntityID
	EntityToIndex   []int
}

func NewComponentManager[T any]() *ComponentManager[T] {
	sparse := make([]int, maxEntityCount)
	for i := range sparse {
		sparse[i] = -1
	}

	return &ComponentManager[T]{
		data:            make([]T, 0, 1024),
		IndexToEntityID: make([]EntityID, 0, 1024),
		EntityToIndex:   sparse,
	}
}

func (cm *ComponentManager[T]) Upsert(entityID EntityID, component T) bool {
	if _, ok := cm.Get(entityID); !ok {
		return cm.Add(entityID, component)
	}

	return cm.Set(entityID, component)
}

func (cm *ComponentManager[T]) Add(entityID EntityID, component T) bool {
	idx := entityID.Index()

	if idx >= len(cm.EntityToIndex) {
		newCap := idx * 2
		if newCap < idx+1 {
			newCap = idx + 1
		}
		newSparse := make([]int, newCap)
		copy(newSparse, cm.EntityToIndex)
		for i := len(cm.EntityToIndex); i < len(newSparse); i++ {
			newSparse[i] = -1
		}
		cm.EntityToIndex = newSparse
	}

	if cm.EntityToIndex[idx] != -1 { // already has component
		return false
	}

	cm.data = append(cm.data, component)
	cm.IndexToEntityID = append(cm.IndexToEntityID, entityID)

	index := len(cm.data) - 1
	cm.EntityToIndex[idx] = index
	return true
}

func (cm *ComponentManager[T]) Get(entityID EntityID) (T, bool) {
	eid := entityID.Index()
	idx := -1
	if eid < len(cm.EntityToIndex) {
		idx = cm.EntityToIndex[eid]
	}

	if idx == -1 {
		var zero T
		return zero, false
	}
	return cm.data[idx], true
}

func (cm *ComponentManager[T]) Set(entityID EntityID, component T) bool {
	eid := entityID.Index()
	if eid >= len(cm.EntityToIndex) {
		return false
	}

	idx := cm.EntityToIndex[eid]
	if idx == -1 {
		return false
	}

	cm.data[idx] = component
	return true
}

func (cm *ComponentManager[T]) Remove(entityID EntityID) bool {
	eid := entityID.Index()
	if eid >= len(cm.EntityToIndex) {
		return false
	}

	idx := cm.EntityToIndex[eid]
	if idx == -1 {
		return false
	}

	lastIndex := len(cm.data) - 1
	if idx != lastIndex {
		lastEntityID := cm.IndexToEntityID[lastIndex]

		cm.data[idx] = cm.data[lastIndex]
		cm.IndexToEntityID[idx] = lastEntityID
		cm.EntityToIndex[lastEntityID.Index()] = idx
	}

	cm.EntityToIndex[eid] = -1
	cm.data = cm.data[:lastIndex]
	cm.IndexToEntityID = cm.IndexToEntityID[:lastIndex]
	return true
}

func (cm *ComponentManager[T]) All() iter.Seq2[EntityID, T] {
	return func(yield func(EntityID, T) bool) {
		for i, component := range cm.data {
			entityID := cm.IndexToEntityID[i]
			if !yield(entityID, component) {
				return
			}
		}
	}
}
