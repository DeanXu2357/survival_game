package state

type ComponentManger[T any] struct {
	data            []T
	IndexToEntityID []EntityID
	EntityToIndex   []int
}

func NewComponentManager[T any]() *ComponentManger[T] {
	sparse := make([]int, maxEntityCount)
	for i := range sparse {
		sparse[i] = -1
	}

	return &ComponentManger[T]{
		data:            make([]T, 0, 1024),
		IndexToEntityID: make([]EntityID, 0, 1024),
		EntityToIndex:   sparse,
	}
}

func (cm *ComponentManger[T]) Add(entityID EntityID, component T) bool {
	// TODO: I should use entityID.Index() as spare index for less slice capacity usage, maybe fix later
	if int(entityID) >= len(cm.EntityToIndex) {
		newCap := int(entityID) * 2
		if newCap < int(entityID)+1 {
			newCap = int(entityID) + 1
		}
		newSparse := make([]int, newCap)
		copy(newSparse, cm.EntityToIndex)
		for i := len(cm.EntityToIndex); i < len(newSparse); i++ {
			newSparse[i] = -1
		}
		cm.EntityToIndex = newSparse
	}

	if cm.EntityToIndex[entityID] != -1 { // already has component
		return false
	}

	cm.data = append(cm.data, component)
	cm.IndexToEntityID = append(cm.IndexToEntityID, entityID)

	index := len(cm.data) - 1
	cm.EntityToIndex[entityID] = index
	return true
}

func (cm *ComponentManger[T]) Get(entityID EntityID) (T, bool) {
	idx := -1
	if int(entityID) < len(cm.EntityToIndex) {
		idx = cm.EntityToIndex[entityID]
	}

	if idx == -1 {
		var zero T
		return zero, false
	}
	return cm.data[idx], true
}

func (cm *ComponentManger[T]) Remove(entityID EntityID) bool {
	idx := cm.EntityToIndex[entityID]
	if idx == -1 {
		return false
	}

	lastIndex := len(cm.data) - 1
	if idx != lastIndex {
		lastEntityID := cm.IndexToEntityID[lastIndex]

		cm.data[idx] = cm.data[lastIndex]
		cm.IndexToEntityID[idx] = lastEntityID
		cm.EntityToIndex[lastEntityID] = idx
	}

	cm.EntityToIndex[entityID] = -1
	cm.data = cm.data[:lastIndex]
	cm.IndexToEntityID = cm.IndexToEntityID[:lastIndex]
	return true
}
