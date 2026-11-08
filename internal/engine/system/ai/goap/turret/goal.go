package turret

import "survival/internal/engine/system/ai/goap"

var _ goap.Goal = (*goal)(nil)

// goal: the turret wants its cone clear of threats.
type goal struct{}

func newGoal() *goal { return &goal{} }

func (g *goal) Require() goap.SymbolSet {
	var s goap.SymbolSet
	return s.With(symTargetDown)
}
