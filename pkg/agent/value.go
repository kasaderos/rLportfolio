package agent

import "github.com/kasaderos/rLportfolio/pkg/rl/state"

// ValueFunction represents a value function (e.g., Q-function).
type ValueFunction interface {
	// Get returns the value for a state-action pair.
	Get(s state.State, a Action) float64
	// Set sets the value for a state-action pair.
	Set(s state.State, a Action, value float64)
	// Max returns the maximum value over actions for a given state.
	Max(s state.State) float64
}

// QTable implements a tabular Q-function.
type QTable struct {
	Q [][]float64 // Q[state][action]
}

// NewQTable creates a new Q-table with the specified dimensions.
func NewQTable(numStates int, numActions int) *QTable {
	Q := make([][]float64, numStates)
	for s := 0; s < numStates; s++ {
		Q[s] = make([]float64, numActions)
	}
	return &QTable{Q: Q}
}

// Get returns the Q-value for a state-action pair.
func (q *QTable) Get(s state.State, a Action) float64 {
	return q.Q[s.Index][int(a)]
}

// Set sets the Q-value for a state-action pair.
func (q *QTable) Set(s state.State, a Action, value float64) {
	q.Q[s.Index][int(a)] = value
}

// Max returns the maximum Q-value over actions for a given state.
func (q *QTable) Max(s state.State) float64 {
	return MaxValue(q.Q[s.Index])
}

// MaxValue returns the maximum value in a slice.
func MaxValue(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	max := arr[0]
	for _, v := range arr[1:] {
		if v > max {
			max = v
		}
	}
	return max
}
