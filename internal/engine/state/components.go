package state

import "survival/internal/engine/vector"

type Position vector.Vector2D

type Direction float64

type PrePosition Position

type MovementSpeed float64

type RotationSpeed float64

type Health int

type ViewIDs []EntityID
