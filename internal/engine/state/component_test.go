package state

import "testing"

type testComponent struct {
	value int
}

func TestNewComponentManager(t *testing.T) {
	cm := NewComponentManager[testComponent]()

	if len(cm.data) != 0 {
		t.Errorf("data should be empty, got len=%d", len(cm.data))
	}
	if len(cm.IndexToEntityID) != 0 {
		t.Errorf("IndexToEntityID should be empty, got len=%d", len(cm.IndexToEntityID))
	}
	if len(cm.EntityToIndex) != maxEntityCount {
		t.Errorf("EntityToIndex should have len=%d, got len=%d", maxEntityCount, len(cm.EntityToIndex))
	}
	for i, idx := range cm.EntityToIndex {
		if idx != -1 {
			t.Errorf("EntityToIndex[%d] should be -1, got %d", i, idx)
			break
		}
	}
}

func TestComponentManager_Add(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*ComponentManager[testComponent])
		entityID EntityID
		comp     testComponent
		wantOK   bool
	}{
		{
			name:     "success",
			setup:    func(cm *ComponentManager[testComponent]) {},
			entityID: NewEntityID(0, 0),
			comp:     testComponent{value: 42},
			wantOK:   true,
		},
		{
			name: "duplicate returns false",
			setup: func(cm *ComponentManager[testComponent]) {
				cm.Add(NewEntityID(0, 0), testComponent{value: 1})
			},
			entityID: NewEntityID(0, 0),
			comp:     testComponent{value: 2},
			wantOK:   false,
		},
		{
			name:     "auto resize for large index",
			setup:    func(cm *ComponentManager[testComponent]) {},
			entityID: NewEntityID(maxEntityCount+100, 0),
			comp:     testComponent{value: 99},
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewComponentManager[testComponent]()
			tt.setup(cm)

			got := cm.Add(tt.entityID, tt.comp)
			if got != tt.wantOK {
				t.Errorf("Add() = %v, want %v", got, tt.wantOK)
			}

			if tt.wantOK {
				stored, ok := cm.Get(tt.entityID)
				if !ok {
					t.Errorf("Get() after Add() should return true")
				}
				if stored.value != tt.comp.value {
					t.Errorf("stored value = %d, want %d", stored.value, tt.comp.value)
				}
			}
		})
	}
}

func TestComponentManager_Get(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*ComponentManager[testComponent])
		entityID  EntityID
		wantValue int
		wantOK    bool
	}{
		{
			name: "found",
			setup: func(cm *ComponentManager[testComponent]) {
				cm.Add(NewEntityID(5, 0), testComponent{value: 100})
			},
			entityID:  NewEntityID(5, 0),
			wantValue: 100,
			wantOK:    true,
		},
		{
			name:      "not found",
			setup:     func(cm *ComponentManager[testComponent]) {},
			entityID:  NewEntityID(10, 0),
			wantValue: 0,
			wantOK:    false,
		},
		{
			name:      "out of bounds",
			setup:     func(cm *ComponentManager[testComponent]) {},
			entityID:  NewEntityID(maxEntityCount+500, 0),
			wantValue: 0,
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewComponentManager[testComponent]()
			tt.setup(cm)

			got, ok := cm.Get(tt.entityID)
			if ok != tt.wantOK {
				t.Errorf("Get() ok = %v, want %v", ok, tt.wantOK)
			}
			if got.value != tt.wantValue {
				t.Errorf("Get() value = %d, want %d", got.value, tt.wantValue)
			}
		})
	}
}

func TestComponentManager_Remove(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*ComponentManager[testComponent])
		entityID EntityID
		wantOK   bool
	}{
		{
			name: "success",
			setup: func(cm *ComponentManager[testComponent]) {
				cm.Add(NewEntityID(0, 0), testComponent{value: 1})
			},
			entityID: NewEntityID(0, 0),
			wantOK:   true,
		},
		{
			name:     "not found",
			setup:    func(cm *ComponentManager[testComponent]) {},
			entityID: NewEntityID(99, 0),
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewComponentManager[testComponent]()
			tt.setup(cm)

			got := cm.Remove(tt.entityID)
			if got != tt.wantOK {
				t.Errorf("Remove() = %v, want %v", got, tt.wantOK)
			}

			if tt.wantOK {
				_, ok := cm.Get(tt.entityID)
				if ok {
					t.Errorf("Get() after Remove() should return false")
				}
			}
		})
	}
}

func TestComponentManager_Remove_SwapBehavior(t *testing.T) {
	cm := NewComponentManager[testComponent]()

	e0 := NewEntityID(0, 0)
	e1 := NewEntityID(1, 0)
	e2 := NewEntityID(2, 0)

	cm.Add(e0, testComponent{value: 10})
	cm.Add(e1, testComponent{value: 20})
	cm.Add(e2, testComponent{value: 30})

	cm.Remove(e0)

	if len(cm.data) != 2 {
		t.Errorf("after remove, data len = %d, want 2", len(cm.data))
	}

	c1, ok1 := cm.Get(e1)
	if !ok1 || c1.value != 20 {
		t.Errorf("e1 should still exist with value 20, got ok=%v, value=%d", ok1, c1.value)
	}

	c2, ok2 := cm.Get(e2)
	if !ok2 || c2.value != 30 {
		t.Errorf("e2 should still exist with value 30, got ok=%v, value=%d", ok2, c2.value)
	}

	_, ok0 := cm.Get(e0)
	if ok0 {
		t.Errorf("e0 should no longer exist after removal")
	}
}

func TestComponentManager_Integration(t *testing.T) {
	cm := NewComponentManager[testComponent]()

	entities := make([]EntityID, 5)
	for i := 0; i < 5; i++ {
		entities[i] = NewEntityID(i, 0)
		cm.Add(entities[i], testComponent{value: (i + 1) * 10})
	}

	for i, e := range entities {
		c, ok := cm.Get(e)
		if !ok {
			t.Errorf("entity %d should exist", i)
		}
		expected := (i + 1) * 10
		if c.value != expected {
			t.Errorf("entity %d value = %d, want %d", i, c.value, expected)
		}
	}

	cm.Remove(entities[2])

	for i, e := range entities {
		c, ok := cm.Get(e)
		if i == 2 {
			if ok {
				t.Errorf("entity 2 should not exist after removal")
			}
		} else {
			if !ok {
				t.Errorf("entity %d should still exist", i)
			}
			expected := (i + 1) * 10
			if c.value != expected {
				t.Errorf("entity %d value = %d, want %d", i, c.value, expected)
			}
		}
	}

	if len(cm.data) != 4 {
		t.Errorf("data len = %d, want 4", len(cm.data))
	}
}

func TestComponentManager_Set(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*ComponentManager[testComponent])
		entityID EntityID
		newComp  testComponent
		wantOK   bool
	}{
		{
			name: "success update",
			setup: func(cm *ComponentManager[testComponent]) {
				cm.Add(NewEntityID(1, 0), testComponent{value: 10})
			},
			entityID: NewEntityID(1, 0),
			newComp:  testComponent{value: 99},
			wantOK:   true,
		},
		{
			name:     "fail - component not added yet",
			setup:    func(cm *ComponentManager[testComponent]) {}, // 空的
			entityID: NewEntityID(1, 0),
			newComp:  testComponent{value: 99},
			wantOK:   false,
		},
		{
			name:     "fail - out of bounds",
			setup:    func(cm *ComponentManager[testComponent]) {},
			entityID: NewEntityID(maxEntityCount+100, 0),
			newComp:  testComponent{value: 99},
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewComponentManager[testComponent]()
			tt.setup(cm)

			got := cm.Set(tt.entityID, tt.newComp)
			if got != tt.wantOK {
				t.Errorf("Set() = %v, want %v", got, tt.wantOK)
			}

			if tt.wantOK {
				stored, _ := cm.Get(tt.entityID)
				if stored.value != tt.newComp.value {
					t.Errorf("stored value = %d, want %d", stored.value, tt.newComp.value)
				}
			}
		})
	}
}

func TestComponentManager_Remove_Boundaries(t *testing.T) {
	t.Run("Sequential Removal to Empty", func(t *testing.T) {
		cm := NewComponentManager[testComponent]()

		e0 := NewEntityID(10, 0) // Index 0 in dense
		e1 := NewEntityID(20, 0) // Index 1 in dense
		e2 := NewEntityID(30, 0) // Index 2 in dense (Last)

		cm.Add(e0, testComponent{value: 100})
		cm.Add(e1, testComponent{value: 200})
		cm.Add(e2, testComponent{value: 300})

		if !cm.Remove(e2) {
			t.Fatal("failed to remove tail element e2")
		}

		if len(cm.data) != 2 {
			t.Errorf("expected len 2, got %d", len(cm.data))
		}
		if _, ok := cm.Get(e2); ok {
			t.Error("e2 should be removed")
		}
		if idx := cm.EntityToIndex[e2.Index()]; idx != -1 {
			t.Errorf("e2 sparse index should be -1, got %d", idx)
		}
		if v, _ := cm.Get(e0); v.value != 100 {
			t.Error("e0 data corrupted")
		}
		if v, _ := cm.Get(e1); v.value != 200 {
			t.Error("e1 data corrupted")
		}

		if !cm.Remove(e1) {
			t.Fatal("failed to remove e1")
		}
		if len(cm.data) != 1 {
			t.Errorf("expected len 1, got %d", len(cm.data))
		}
		if _, ok := cm.Get(e1); ok {
			t.Error("e1 should be removed")
		}
		if v, _ := cm.Get(e0); v.value != 100 {
			t.Error("e0 data corrupted")
		}

		if !cm.Remove(e0) {
			t.Fatal("failed to remove last standing e0")
		}
		if len(cm.data) != 0 {
			t.Errorf("expected len 0, got %d", len(cm.data))
		}
		if len(cm.IndexToEntityID) != 0 {
			t.Errorf("expected IndexToEntityID empty, got %d", len(cm.IndexToEntityID))
		}
		if _, ok := cm.Get(e0); ok {
			t.Error("e0 should be removed")
		}
	})

	t.Run("Invalid Operations", func(t *testing.T) {
		cm := NewComponentManager[testComponent]()
		e1 := NewEntityID(1, 0)
		cm.Add(e1, testComponent{value: 10})

		tests := []struct {
			name     string
			entityID EntityID
			want     bool
		}{
			{
				name:     "Remove non-existent entity",
				entityID: NewEntityID(99, 0),
				want:     false,
			},
			{
				name:     "Remove out of bounds entity",
				entityID: NewEntityID(maxEntityCount+500, 0),
				want:     false,
			},
			{
				name:     "Double remove (Idempotency check)",
				entityID: e1,
				want:     false,
			},
		}

		if !cm.Remove(e1) {
			t.Fatal("Setup failed: could not remove e1")
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := cm.Remove(tt.entityID)
				if got != tt.want {
					t.Errorf("Remove() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}
