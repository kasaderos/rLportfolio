package agent

import (
	"math/rand"

	"github.com/kasaderos/rLportfolio/pkg/state"
)

// Actor defines the policy interface for selecting actions.
type Actor interface {
	Act(s state.State) Action
}

// Policy represents a policy that can select actions and be updated.
type Policy interface {
	Actor
	// SetExploration sets the exploration rate (epsilon for epsilon-greedy)
	SetExploration(epsilon float64)
}

// EpsilonGreedyPolicy implements epsilon-greedy action selection.
type EpsilonGreedyPolicy struct {
	Q       [][]float64 // Q-table: Q[state][action]
	Epsilon float64     // Exploration rate
	RNG     *rand.Rand
}

// NewEpsilonGreedyPolicy creates a new epsilon-greedy policy.
func NewEpsilonGreedyPolicy(Q [][]float64, epsilon float64, rng *rand.Rand) *EpsilonGreedyPolicy {
	return &EpsilonGreedyPolicy{
		Q:       Q,
		Epsilon: epsilon,
		RNG:     rng,
	}
}

// Act selects an action using epsilon-greedy strategy.
func (p *EpsilonGreedyPolicy) Act(s state.State) Action {
	if p.RNG.Float64() < p.Epsilon {
		// Explore: random action
		return Action(p.RNG.Intn(int(NumActions)))
	}
	// Exploit: best action according to Q-table
	return p.greedyAction(s)
}

// SetExploration sets the exploration rate.
func (p *EpsilonGreedyPolicy) SetExploration(epsilon float64) {
	p.Epsilon = epsilon
}

// greedyAction returns the action with highest Q-value for the state.
func (p *EpsilonGreedyPolicy) greedyAction(s state.State) Action {
	return Action(ArgMax(p.Q[s.Index]))
}

// ArgMax returns the index of the maximum value in the slice.
func ArgMax(arr []float64) int {
	if len(arr) == 0 {
		return 0
	}
	best := 0
	for i := 1; i < len(arr); i++ {
		if arr[i] > arr[best] {
			best = i
		}
	}
	return best
}

// GreedyPolicy is a policy that always selects the best action (no exploration).
type GreedyPolicy struct {
	Q [][]float64
}

// NewGreedyPolicy creates a new greedy policy.
func NewGreedyPolicy(Q [][]float64) *GreedyPolicy {
	return &GreedyPolicy{Q: Q}
}

// Act selects the best action according to Q-table.
func (p *GreedyPolicy) Act(s state.State) Action {
	return Action(ArgMax(p.Q[s.Index]))
}

// SetExploration is a no-op for greedy policy.
func (p *GreedyPolicy) SetExploration(epsilon float64) {}
