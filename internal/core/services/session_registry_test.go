package services

import (
	"sync"
	"testing"

	"survival/internal/core/domain/state"
)

func TestSessionRegistry_RegisterAndEntityID(t *testing.T) {
	sr := NewSessionRegistry()
	sessionID := "session-123"
	entityID := state.EntityID(42)

	sr.Register(sessionID, entityID)

	got, ok := sr.EntityID(sessionID)
	if !ok {
		t.Fatal("Expected to find registered session")
	}
	if got != entityID {
		t.Errorf("EntityID() = %d, want %d", got, entityID)
	}
}

func TestSessionRegistry_RegisterAndSessionID(t *testing.T) {
	sr := NewSessionRegistry()
	sessionID := "session-123"
	entityID := state.EntityID(42)

	sr.Register(sessionID, entityID)

	got, ok := sr.SessionID(entityID)
	if !ok {
		t.Fatal("Expected to find registered entity")
	}
	if got != sessionID {
		t.Errorf("SessionID() = %s, want %s", got, sessionID)
	}
}

func TestSessionRegistry_EntityID_NotFound(t *testing.T) {
	sr := NewSessionRegistry()

	_, ok := sr.EntityID("nonexistent")
	if ok {
		t.Error("Expected not to find unregistered session")
	}
}

func TestSessionRegistry_SessionID_NotFound(t *testing.T) {
	sr := NewSessionRegistry()

	_, ok := sr.SessionID(state.EntityID(999))
	if ok {
		t.Error("Expected not to find unregistered entity")
	}
}

func TestSessionRegistry_Unregister(t *testing.T) {
	sr := NewSessionRegistry()
	sessionID := "session-123"
	entityID := state.EntityID(42)

	sr.Register(sessionID, entityID)
	sr.Unregister(sessionID)

	_, ok := sr.EntityID(sessionID)
	if ok {
		t.Error("Expected session to be unregistered")
	}

	_, ok = sr.SessionID(entityID)
	if ok {
		t.Error("Expected entity to be unregistered")
	}
}

func TestSessionRegistry_Unregister_NonExistent(t *testing.T) {
	sr := NewSessionRegistry()
	sr.Unregister("nonexistent")
}

func TestSessionRegistry_AllSessionIDs(t *testing.T) {
	sr := NewSessionRegistry()

	sr.Register("session-1", state.EntityID(1))
	sr.Register("session-2", state.EntityID(2))
	sr.Register("session-3", state.EntityID(3))

	ids := sr.AllSessionIDs()
	if len(ids) != 3 {
		t.Errorf("AllSessionIDs() returned %d items, want 3", len(ids))
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	for _, expected := range []string{"session-1", "session-2", "session-3"} {
		if !idSet[expected] {
			t.Errorf("AllSessionIDs() missing %s", expected)
		}
	}
}

func TestSessionRegistry_AllSessionIDs_Empty(t *testing.T) {
	sr := NewSessionRegistry()
	ids := sr.AllSessionIDs()
	if len(ids) != 0 {
		t.Errorf("AllSessionIDs() on empty registry returned %d items, want 0", len(ids))
	}
}

func TestSessionRegistry_Clear(t *testing.T) {
	sr := NewSessionRegistry()

	sr.Register("session-1", state.EntityID(1))
	sr.Register("session-2", state.EntityID(2))

	sr.Clear()

	if sr.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", sr.Count())
	}

	_, ok := sr.EntityID("session-1")
	if ok {
		t.Error("Expected session-1 to be cleared")
	}
}

func TestSessionRegistry_Count(t *testing.T) {
	sr := NewSessionRegistry()

	if sr.Count() != 0 {
		t.Errorf("Count() on new registry = %d, want 0", sr.Count())
	}

	sr.Register("session-1", state.EntityID(1))
	if sr.Count() != 1 {
		t.Errorf("Count() after 1 register = %d, want 1", sr.Count())
	}

	sr.Register("session-2", state.EntityID(2))
	if sr.Count() != 2 {
		t.Errorf("Count() after 2 registers = %d, want 2", sr.Count())
	}

	sr.Unregister("session-1")
	if sr.Count() != 1 {
		t.Errorf("Count() after unregister = %d, want 1", sr.Count())
	}
}

func TestSessionRegistry_ConcurrentAccess(t *testing.T) {
	sr := NewSessionRegistry()
	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines * 3)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			sr.Register("session-"+string(rune('a'+i%26)), state.EntityID(i))
		}(i)

		go func(i int) {
			defer wg.Done()
			sr.EntityID("session-" + string(rune('a'+i%26)))
		}(i)

		go func(i int) {
			defer wg.Done()
			sr.AllSessionIDs()
		}(i)
	}

	wg.Wait()
}

func TestSessionRegistry_OverwriteRegistration(t *testing.T) {
	sr := NewSessionRegistry()

	sr.Register("session-1", state.EntityID(100))
	sr.Register("session-1", state.EntityID(200))

	got, ok := sr.EntityID("session-1")
	if !ok {
		t.Fatal("Expected to find session after overwrite")
	}
	if got != 200 {
		t.Errorf("EntityID() after overwrite = %d, want 200", got)
	}

	_, ok = sr.SessionID(state.EntityID(100))
	if ok {
		t.Error("Old entity ID should no longer be mapped (stale mapping)")
	}
}
