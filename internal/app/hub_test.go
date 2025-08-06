package app

import (
	"context"
	"testing"

	"survival/internal/utils"
)

func TestNewHub_WithMapLoading(t *testing.T) {
	ctx := context.Background()
	idGen := utils.NewSequentialIDGenerator("test")

	// Create hub - should try to load office_floor_01, fallback to embedded default
	hub := NewHub(ctx, idGen)
	defer hub.Shutdown(ctx)

	// Verify hub was created successfully
	if hub == nil {
		t.Fatal("Hub should not be nil")
	}

	// Verify default room exists
	if len(hub.rooms) != 1 {
		t.Errorf("Expected 1 room, got %d", len(hub.rooms))
	}

	room, exists := hub.rooms[DefaultRoomName]
	if !exists {
		t.Error("Default room should exist")
	}

	if room == nil {
		t.Error("Default room should not be nil")
	}
}
