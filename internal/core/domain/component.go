package domain

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

	dataIdx := len(cm.data) - 1
	cm.EntityToIndex[idx] = dataIdx
	return true
}

func (cm *ComponentManger[T]) Get(entityID EntityID) *T {
	entityIdx := entityID.Index()
	if entityIdx >= len(cm.EntityToIndex) {
		return nil
	}
	dataIdx := cm.EntityToIndex[entityIdx]
	if dataIdx == -1 {
		return nil
	}
	// Verify version matches to prevent accessing stale data
	if cm.IndexToEntityID[dataIdx] != entityID {
		return nil
	}
	return &cm.data[dataIdx]
}

func (cm *ComponentManger[T]) Remove(entityID EntityID) bool {
	entityIdx := entityID.Index()
	if entityIdx >= len(cm.EntityToIndex) {
		return false
	}
	dataIdx := cm.EntityToIndex[entityIdx]
	if dataIdx == -1 {
		return false
	}
	// Verify version matches to prevent removing wrong entity's component
	if cm.IndexToEntityID[dataIdx] != entityID {
		return false
	}

	lastIndex := len(cm.data) - 1
	if dataIdx != lastIndex {
		lastEntityID := cm.IndexToEntityID[lastIndex]

		cm.data[dataIdx] = cm.data[lastIndex]
		cm.IndexToEntityID[dataIdx] = lastEntityID
		cm.EntityToIndex[lastEntityID.Index()] = dataIdx
	}

	cm.EntityToIndex[entityIdx] = -1
	cm.data = cm.data[:lastIndex]
	cm.IndexToEntityID = cm.IndexToEntityID[:lastIndex]
	return true
}
