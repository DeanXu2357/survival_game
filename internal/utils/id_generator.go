package utils

import (
	"fmt"
	"sync/atomic"
)

// SequentialIDGenerator generates IDs using a prefix and incrementing sequence number
type SequentialIDGenerator struct {
	prefix  string
	counter int64
}

// NewSequentialIDGenerator creates a new sequential ID generator with the given prefix
func NewSequentialIDGenerator(prefix string) *SequentialIDGenerator {
	return &SequentialIDGenerator{
		prefix:  prefix,
		counter: 0,
	}
}

// GenerateID generates a new unique ID with format: prefix-sequenceNumber
func (g *SequentialIDGenerator) GenerateID() string {
	seq := atomic.AddInt64(&g.counter, 1)
	return fmt.Sprintf("%s-%d", g.prefix, seq)
}

// GetPrefix returns the prefix of the ID generator
func (g *SequentialIDGenerator) GetPrefix() string {
	return g.prefix
}