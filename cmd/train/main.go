package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/kasaderos/rLportfolio/pkg/rl/agent"
	"github.com/kasaderos/rLportfolio/pkg/rl/env"
	"github.com/kasaderos/rLportfolio/pkg/rl/plot"
	"github.com/kasaderos/rLportfolio/pkg/rl/state"
	"github.com/kasaderos/rLportfolio/pkg/rl/trainer"
)

const (
	// Q-learning parameters
	alpha   = 0.1
	gamma   = 0.95
	epsilon = 0.1

	episodes  = 1000
	minPrices = 50 // Minimum prices needed to start training

	// Local approximation parameters
	approxM1 = 7
	approxN1 = 7
	approxM2 = 14
	approxN2 = 7
)

func main() {
	seed := flag.Int64("seed", time.Now().UnixNano(), "random seed")
	seriesLength := flag.Int("series-length", 1000, "series length")
	episodeCount := flag.Int("episode-count", 0, "episode count")
	flag.Parse()

	if *episodeCount <= 0 {
		*episodeCount = episodes
	}
	if *seriesLength <= 0 {
		*seriesLength = 1000
	}

	rng := rand.New(rand.NewSource(*seed))

	// Generate synthetic prices
	prices := generateSyntheticPrices(rng, *seriesLength)
	if len(prices) < minPrices {
		fmt.Printf("Error: Need at least %d prices, got %d\n", minPrices, len(prices))
		return
	}

	// Create environment
	marketEnv := env.NewMarketEnv(env.MarketConfig{
		Prices:      prices,
		InitialCash: 10000.0,
		ApproxM1:    approxM1,
		ApproxN1:    approxN1,
		ApproxM2:    approxM2,
		ApproxN2:    approxN2,
		MinStartIdx: 20,
	})

	// Create Q-table and policy
	Q := agent.NewQTable(state.NumStates, agent.NumActions)
	policy := agent.NewEpsilonGreedyPolicy(Q.Q, epsilon, rng)

	// Create agent
	rlAgent := agent.NewQLearningAgent(Q, policy, alpha, gamma)

	// Create trainer
	t := trainer.NewTrainer(marketEnv, rlAgent)

	fmt.Printf("Starting Q-learning training with %d prices...\n", len(prices))
	fmt.Printf("Initial portfolio: Cash=%.2f, Shares=%.2f\n\n", marketEnv.Cash(), marketEnv.Shares())

	// Train
	t.Run(*episodeCount, 100)

	// Extract and display policy
	fmt.Println("\n=== Optimal Policy ===")
	displayPolicy(Q.Q)

	// Test the learned policy and get series data
	fmt.Println("\n=== Testing Learned Policy ===")
	cashSeries, actions := testPolicy(Q.Q, prices, marketEnv)

	// Save series data to rl/plot/data/series.csv
	if err := plot.SaveSeriesData(prices, cashSeries, actions); err != nil {
		fmt.Printf("Failed to save series: %v\n", err)
	} else {
		fmt.Println("Saved series data to rl/plot/data/series.csv")
	}

	// Save Q-matrix to rl/plot/data/q_matrix.csv
	if err := plot.SaveQMatrixData(Q.Q); err != nil {
		fmt.Printf("Failed to save Q matrix: %v\n", err)
	} else {
		fmt.Println("Saved Q matrix to rl/plot/data/q_matrix.csv")
	}
}

// displayPolicy displays the optimal policy for all states.
func displayPolicy(Q [][]float64) {
	for expRet1 := 0; expRet1 < state.NumExpRetCategories; expRet1++ {
		for dist1 := 0; dist1 < state.NumMinDistCategories; dist1++ {
			for expRet2 := 0; expRet2 < state.NumExpRetCategories; expRet2++ {
				for dist2 := 0; dist2 < state.NumMinDistCategories; dist2++ {
					s := state.NewState(expRet1, dist1, expRet2, dist2)
					action := agent.ArgMax(Q[s.Index])
					fmt.Printf("State (expRet1=%d, dist1=%d, expRet2=%d, dist2=%d) -> Action %d (%s)\n",
						expRet1, dist1, expRet2, dist2, action, agent.Action(action).String())
				}
			}
		}
	}
}

// testPolicy tests the learned policy on the price data and returns cash series and actions.
func testPolicy(Q [][]float64, prices []float64, marketEnv *env.MarketEnv) ([]float64, []int) {
	// Create greedy policy for testing
	greedyPolicy := agent.NewGreedyPolicy(Q)
	testAgent := &testAgent{policy: greedyPolicy}

	// Reset environment
	s := marketEnv.Reset()
	done := false
	actions := make([]int, len(prices))
	cashSeries := make([]float64, len(prices))
	initialValue := marketEnv.PortfolioValue()

	for i := range actions {
		actions[i] = -1
		cashSeries[i] = marketEnv.Cash()
	}

	step := 0
	for !done {
		action := testAgent.Act(s)
		next, _, d := marketEnv.Step(action)
		actions[step] = int(action)
		cashSeries[step+1] = marketEnv.Cash()
		s = next
		done = d
		step++
	}

	finalValue := marketEnv.PortfolioValue()
	returnPct := (finalValue/initialValue - 1.0) * 100

	fmt.Printf("Test Results:\n")
	fmt.Printf("  Initial value: %.2f\n", initialValue)
	fmt.Printf("  Final value: %.2f\n", finalValue)
	fmt.Printf("  Return: %.2f%%\n", returnPct)
	fmt.Printf("  Final cash: %.2f\n", marketEnv.Cash())
	fmt.Printf("  Final shares: %.2f\n", marketEnv.Shares())

	if err := plot.SavePlots(prices, cashSeries, actions); err != nil {
		fmt.Printf("Failed to save plots: %v\n", err)
	} else {
		fmt.Println("Saved plots: policy_actions.png, cash_series.png")
	}

	return cashSeries, actions
}

// testAgent is a simple agent that only acts (for testing).
type testAgent struct {
	policy agent.Actor
}

func (a *testAgent) Act(s state.State) agent.Action {
	return a.policy.Act(s)
}

// generateSyntheticPrices creates synthetic price data for testing.
func generateSyntheticPrices(rng *rand.Rand, n int) []float64 {
	prices := make([]float64, n)
	prices[0] = 100.0

	for i := 1; i < n; i++ {
		// Random walk with occasional outliers and sharp drops.
		// Base move: slight positive drift with moderate volatility.
		change := (rng.Float64()-0.5)*2.0 + 0.2 // [-0.8%, +1.2%]

		roll := rng.Float64()
		switch {
		case roll < 0.02:
			// Sharp drop: -8% to -25%
			change = -8.0 - rng.Float64()*17.0
		case roll < 0.04:
			// Outlier jump: +8% to +20%
			change = 8.0 + rng.Float64()*12.0
		case roll < 0.06:
			// Mild drop cluster: -2% to -6%
			change = -2.0 - rng.Float64()*4.0
		case roll < 0.12:
			// Mild growth cluster: +2% to +6%
			change = 2.0 + rng.Float64()*4.0
		}

		next := prices[i-1] * (1.0 + change/100.0)
		if next <= 0 {
			next = prices[i-1] * 0.5
		}
		prices[i] = next
	}

	return prices
}
