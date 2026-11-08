package turret

import "survival/internal/engine/system/ai/goap"

var _ goap.Planner = (*planner)(nil)

// planner is a forward-search A* over symbol-set states. The turret's action and
// symbol space is tiny, so a plain best-first expansion with a per-state best
// cost is sufficient and always terminates.
type planner struct{}

func newPlanner() *planner { return &planner{} }

type searchNode struct {
	state goap.SymbolSet
	plan  []goap.ActionOnSymbolic
	cost  float64
}

func (p *planner) Plan(start goap.SymbolSet, goal goap.Goal, actions []goap.ActionOnSymbolic) ([]goap.ActionOnSymbolic, bool) {
	want := goal.Require()
	if start.Contains(want) {
		return nil, true
	}

	bestCost := map[goap.SymbolSet]float64{start: 0}
	frontier := []searchNode{{state: start}}

	var best []goap.ActionOnSymbolic
	found := false
	var bestPlanCost float64

	for len(frontier) > 0 {
		idx := 0
		for i := 1; i < len(frontier); i++ {
			if frontier[i].cost+heuristic(frontier[i].state, want) < frontier[idx].cost+heuristic(frontier[idx].state, want) {
				idx = i
			}
		}
		cur := frontier[idx]
		frontier = append(frontier[:idx], frontier[idx+1:]...)

		if found && cur.cost >= bestPlanCost {
			continue
		}

		for _, act := range actions {
			if !cur.state.Contains(act.Preconditions()) {
				continue
			}
			set, clear := act.Effect()
			next := cur.state.Apply(set, clear)
			nextCost := cur.cost + act.StaticCost()
			if prev, seen := bestCost[next]; seen && nextCost >= prev {
				continue
			}
			bestCost[next] = nextCost

			nextPlan := make([]goap.ActionOnSymbolic, len(cur.plan)+1)
			copy(nextPlan, cur.plan)
			nextPlan[len(cur.plan)] = act

			if next.Contains(want) {
				if !found || nextCost < bestPlanCost {
					best = nextPlan
					bestPlanCost = nextCost
					found = true
				}
				continue
			}
			frontier = append(frontier, searchNode{state: next, plan: nextPlan, cost: nextCost})
		}
	}
	return best, found
}

func heuristic(s, want goap.SymbolSet) float64 {
	return float64(want.Missing(s).Count())
}
