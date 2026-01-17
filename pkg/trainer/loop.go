package trainer

import (
	"fmt"

	"github.com/kasaderos/rLportfolio/pkg/agent"
	"github.com/kasaderos/rLportfolio/pkg/env"
)

// Trainer runs training episodes for an RL agent.
type Trainer struct {
	Env   env.Environment
	Agent agent.Agent
}

// NewTrainer creates a new trainer.
func NewTrainer(env env.Environment, agent agent.Agent) *Trainer {
	return &Trainer{
		Env:   env,
		Agent: agent,
	}
}

// Run executes training episodes.
func (t *Trainer) Run(episodes int, reportInterval int) {
	if reportInterval <= 0 {
		reportInterval = 100
	}

	for ep := 0; ep < episodes; ep++ {
		s := t.Env.Reset()
		done := false
		episodeReward := 0.0

		for !done {
			action := t.Agent.Act(s)
			next, reward, d := t.Env.Step(action)

			t.Agent.Learn(agent.Transition{
				State:     s,
				Action:    action,
				Reward:    reward,
				NextState: next,
				Done:      d,
			})

			s = next
			done = d
			episodeReward += reward
		}

		if (ep+1)%reportInterval == 0 {
			// Get final portfolio value if environment supports it
			if marketEnv, ok := t.Env.(*env.MarketEnv); ok {
				finalValue := marketEnv.PortfolioValue()
				initialValue := marketEnv.InitialValue()
				returnPct := (finalValue/initialValue - 1.0) * 100
				fmt.Printf("Episode %d: Final value=%.2f, Return=%.2f%%, Reward=%.4f\n",
					ep+1, finalValue, returnPct, episodeReward)
			} else {
				fmt.Printf("Episode %d: Reward=%.4f\n", ep+1, episodeReward)
			}
		}
	}
}
