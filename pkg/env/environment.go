package env

import (
	"github.com/kasaderos/rLportfolio/pkg/rl/agent"
	"github.com/kasaderos/rLportfolio/pkg/rl/state"
)

// Environment defines the interface for RL environments.
type Environment interface {
	// Reset resets the environment and returns the initial state.
	Reset() state.State
	// Step executes an action and returns the next state, reward, and done flag.
	Step(action agent.Action) (next state.State, reward float64, done bool)
}
