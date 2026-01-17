package agent

import "github.com/kasaderos/rLportfolio/pkg/state"

// Transition represents a state-action-reward-nextState transition.
type Transition struct {
	State     state.State
	Action    Action
	Reward    float64
	NextState state.State
	Done      bool
}
