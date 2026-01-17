package agent

import "github.com/kasaderos/rLportfolio/pkg/rl/state"

// Learner defines the interface for learning from transitions.
type Learner interface {
	Learn(t Transition)
}

// Agent combines Actor and Learner interfaces.
type Agent interface {
	Actor
	Learner
}

// QLearningAgent implements Q-learning algorithm.
type QLearningAgent struct {
	Q      ValueFunction
	Policy Policy
	Alpha  float64 // Learning rate
	Gamma  float64 // Discount factor
}

// NewQLearningAgent creates a new Q-learning agent.
func NewQLearningAgent(Q ValueFunction, policy Policy, alpha, gamma float64) *QLearningAgent {
	return &QLearningAgent{
		Q:      Q,
		Policy: policy,
		Alpha:  alpha,
		Gamma:  gamma,
	}
}

// Act selects an action using the policy.
func (a *QLearningAgent) Act(s state.State) Action {
	return a.Policy.Act(s)
}

// Learn updates the Q-function using Q-learning TD update.
func (a *QLearningAgent) Learn(t Transition) {
	// Current Q-value
	qCurrent := a.Q.Get(t.State, t.Action)

	// TD target: r + gamma * max_a' Q(s', a')
	var qNext float64
	if !t.Done {
		qNext = a.Q.Max(t.NextState)
	}
	tdTarget := t.Reward + a.Gamma*qNext

	// TD error
	tdError := tdTarget - qCurrent

	// Q-learning update: Q(s,a) = Q(s,a) + alpha * (tdTarget - Q(s,a))
	newValue := qCurrent + a.Alpha*tdError
	a.Q.Set(t.State, t.Action, newValue)
}
