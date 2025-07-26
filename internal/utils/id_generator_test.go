package utils

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewSequentialIDGenerator(t *testing.T) {
	prefix := "test"
	generator := NewSequentialIDGenerator(prefix)

	if generator == nil {
		t.Fatal("NewSequentialIDGenerator returned nil")
	}

	if generator.GetPrefix() != prefix {
		t.Errorf("GetPrefix() = %q, want %q", generator.GetPrefix(), prefix)
	}

	if generator.counter != 0 {
		t.Errorf("initial counter = %d, want 0", generator.counter)
	}
}

func TestSequentialIDGenerator_GenerateID_Sequential(t *testing.T) {
	prefix := "player"
	generator := NewSequentialIDGenerator(prefix)

	for i := 1; i <= 3; i++ {
		expectedID := fmt.Sprintf("%s-%d", prefix, i)
		generatedID := generator.GenerateID()
		if generatedID != expectedID {
			t.Errorf("GenerateID() = %q, want %q", generatedID, expectedID)
		}
	}
}

func TestSequentialIDGenerator_GenerateID_Concurrent(t *testing.T) {
	prefix := "concurrent"
	generator := NewSequentialIDGenerator(prefix)
	numGoroutines := 100
	idsToGeneratePerGouroutine := 10

	var wg sync.WaitGroup
	generatedIDs := make(chan string, numGoroutines*idsToGeneratePerGouroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsToGeneratePerGouroutine; j++ {
				generatedIDs <- generator.GenerateID()
			}
		}()
	}

	wg.Wait()
	close(generatedIDs)

	idSet := make(map[string]struct{})
	for id := range generatedIDs {
		if _, exists := idSet[id]; exists {
			t.Errorf("Duplicate ID generated: %q", id)
		}
		idSet[id] = struct{}{}
	}

	expectedNumIDs := numGoroutines * idsToGeneratePerGouroutine
	if len(idSet) != expectedNumIDs {
		t.Errorf("Expected %d unique IDs, but got %d", expectedNumIDs, len(idSet))
	}
}
