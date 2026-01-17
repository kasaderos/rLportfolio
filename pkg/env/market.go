package env

import (
	localapproximation "github.com/kasaderos/rLportfolio/pkg/local-approximation"
	"github.com/kasaderos/rLportfolio/pkg/rl/agent"
	"github.com/kasaderos/rLportfolio/pkg/rl/state"
)

// MarketEnv implements a market trading environment for portfolio optimization.
type MarketEnv struct {
	prices       []float64
	returns      []float64
	currentIdx   int
	cash         float64
	shares       float64
	initialValue float64
	startIdx     int
	approxM1     int
	approxN1     int
	approxM2     int
	approxN2     int
}

// MarketConfig holds configuration for the market environment.
type MarketConfig struct {
	Prices      []float64
	InitialCash float64
	ApproxM1    int
	ApproxN1    int
	ApproxM2    int
	ApproxN2    int
	MinStartIdx int
}

// NewMarketEnv creates a new market environment.
func NewMarketEnv(config MarketConfig) *MarketEnv {
	if config.InitialCash <= 0 {
		config.InitialCash = 10000.0
	}
	if config.MinStartIdx < 1 {
		config.MinStartIdx = 20
	}

	// Calculate returns
	returns := simpleReturns(config.Prices)

	// Determine start index (need enough history for local approximation)
	startIdx := config.ApproxM2 + 1
	if startIdx < config.MinStartIdx {
		startIdx = config.MinStartIdx
	}

	return &MarketEnv{
		prices:       config.Prices,
		returns:      returns,
		currentIdx:   startIdx,
		cash:         config.InitialCash,
		shares:       0.0,
		initialValue: config.InitialCash,
		startIdx:     startIdx,
		approxM1:     config.ApproxM1,
		approxN1:     config.ApproxN1,
		approxM2:     config.ApproxM2,
		approxN2:     config.ApproxN2,
	}
}

// Reset resets the environment to the initial state.
func (e *MarketEnv) Reset() state.State {
	e.currentIdx = e.startIdx
	e.cash = e.initialValue
	e.shares = 0.0
	return e.getState()
}

// Step executes an action and returns the next state, reward, and done flag.
func (e *MarketEnv) Step(action agent.Action) (next state.State, reward float64, done bool) {
	if e.currentIdx >= len(e.prices)-1 {
		return e.getState(), 0.0, true
	}

	currentPrice := e.prices[e.currentIdx]
	nextPrice := e.prices[e.currentIdx+1]

	// Execute action and calculate reward
	portfolioValueBefore := e.cash + e.shares*currentPrice
	e.executeAction(action, currentPrice)
	portfolioValueAfter := e.cash + e.shares*nextPrice
	reward = CalculateReward(portfolioValueBefore, portfolioValueAfter)

	// Move to next time step
	e.currentIdx++

	// Check if done
	done = e.currentIdx >= len(e.prices)-1

	// Get next state
	next = e.getState()

	return next, reward, done
}

// getState computes the current state using local approximation.
func (e *MarketEnv) getState() state.State {
	if e.currentIdx < e.startIdx || e.currentIdx >= len(e.prices) {
		// Return a default state if we don't have enough data
		return state.NewState(0, 0, 0, 0)
	}

	priceHistory := e.prices[:e.currentIdx+1]
	returnsHistory := simpleReturns(priceHistory)
	if len(returnsHistory) < e.approxM2+1 {
		return state.NewState(0, 0, 0, 0)
	}

	// Local approximation call 1
	pred1, minDist1, err := localapproximation.LocalApproximation(returnsHistory, e.approxM1, e.approxN1)
	if err != nil {
		return state.NewState(0, 0, 0, 0)
	}
	expRetCat1 := state.GetExpRetCategory(pred1)
	minDistCat1 := state.GetMinDistCategory(minDist1)

	// Local approximation call 2
	pred2, minDist2, err := localapproximation.LocalApproximation(returnsHistory, e.approxM2, e.approxN2)
	if err != nil {
		return state.NewState(0, 0, 0, 0)
	}
	expRetCat2 := state.GetExpRetCategory(pred2)
	minDistCat2 := state.GetMinDistCategory(minDist2)

	return state.NewState(expRetCat1, minDistCat1, expRetCat2, minDistCat2)
}

// executeAction executes the action and updates cash and shares.
func (e *MarketEnv) executeAction(action agent.Action, price float64) {
	switch action {
	case agent.ActionNothing:
		// No action
	case agent.ActionBuySmall:
		cost := e.cash * agent.BuySmall
		e.cash -= cost
		e.shares += cost / price
	case agent.ActionBuyMedium:
		cost := e.cash * agent.BuyMedium
		e.cash -= cost
		e.shares += cost / price
	case agent.ActionBuyLarge:
		cost := e.cash * agent.BuyLarge
		e.cash -= cost
		e.shares += cost / price
	case agent.ActionSellSmall:
		sellShares := e.shares * agent.SellSmall
		e.cash += sellShares * price
		e.shares -= sellShares
	case agent.ActionSellMedium:
		sellShares := e.shares * agent.SellMedium
		e.cash += sellShares * price
		e.shares -= sellShares
	case agent.ActionSellLarge:
		sellShares := e.shares * agent.SellLarge
		e.cash += sellShares * price
		e.shares -= sellShares
	}
}

// PortfolioValue returns the current portfolio value.
func (e *MarketEnv) PortfolioValue() float64 {
	if e.currentIdx >= len(e.prices) {
		return e.cash
	}
	return e.cash + e.shares*e.prices[e.currentIdx]
}

// Cash returns the current cash amount.
func (e *MarketEnv) Cash() float64 {
	return e.cash
}

// Shares returns the current number of shares.
func (e *MarketEnv) Shares() float64 {
	return e.shares
}

// CurrentPrice returns the current price.
func (e *MarketEnv) CurrentPrice() float64 {
	if e.currentIdx >= len(e.prices) {
		return 0.0
	}
	return e.prices[e.currentIdx]
}

// InitialValue returns the initial portfolio value.
func (e *MarketEnv) InitialValue() float64 {
	return e.initialValue
}

// simpleReturns calculates simple returns from price series.
func simpleReturns(prices []float64) []float64 {
	if len(prices) < 2 {
		return nil
	}
	r := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		r[i-1] = prices[i]/prices[i-1] - 1.0
	}
	return r
}
