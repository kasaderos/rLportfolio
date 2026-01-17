package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kasaderos/rLportfolio/pkg/agent"
	"github.com/kasaderos/rLportfolio/pkg/env"
	"github.com/kasaderos/rLportfolio/pkg/plot"
	"github.com/kasaderos/rLportfolio/pkg/state"
)

const (
	// Local approximation parameters (must match training)
	approxM = 7
	approxN = 5
)

func main() {
	// Load Q-matrix from data/q_matrix.csv
	fmt.Println("Loading Q-matrix from data/q_matrix.csv...")
	Q, err := plot.LoadQMatrixData()
	if err != nil {
		fmt.Printf("Error loading Q-matrix: %v\n", err)
		return
	}
	fmt.Printf("Loaded Q-matrix with %d states and %d actions\n", len(Q), len(Q[0]))

	// Load test prices from data/test.csv
	fmt.Println("\nLoading test prices from data/test.csv...")
	prices, err := loadTestPricesFromCSV("data/test.csv")
	if err != nil {
		fmt.Printf("Error loading test prices: %v\n", err)
		return
	}
	if len(prices) < 50 {
		fmt.Printf("Error: Need at least 50 prices, got %d\n", len(prices))
		return
	}
	fmt.Printf("Loaded %d test prices\n", len(prices))

	// Create market environment with test prices
	marketEnv := env.NewMarketEnv(env.MarketConfig{
		Prices:      prices,
		InitialCash: 10000.0,
		ApproxM:     approxM,
		ApproxN:     approxN,
		MinStartIdx: 120,   // Need at least 120 for MA120
		Commission:  0.002, // 2% commission
	})

	fmt.Printf("Initial portfolio: Cash=%.2f, Shares=%.2f\n\n", marketEnv.Cash(), marketEnv.Shares())

	// Test the learned policy on test data
	fmt.Println("=== Testing Learned Policy on Test Data ===")
	portfolioSeries, actions, actionData := testPolicy(Q, prices, marketEnv)

	// Save test series data to data/test_series.csv
	fmt.Println("\nSaving test results to data/test_series.csv...")
	if err := plot.SaveSeriesDataToFile(prices, portfolioSeries, actions, actionData, "data/test_series.csv"); err != nil {
		fmt.Printf("Failed to save test series: %v\n", err)
		return
	}

	fmt.Println("Test series data saved to data/test_series.csv")
}

// testPolicy tests the learned policy on the price data and returns portfolio value series, actions, and action data.
func testPolicy(Q [][]float64, prices []float64, marketEnv *env.MarketEnv) ([]float64, []int, []plot.ActionData) {
	// Create greedy policy for testing
	greedyPolicy := agent.NewGreedyPolicy(Q)
	testAgent := &testAgent{policy: greedyPolicy}

	// Reset environment
	s := marketEnv.Reset()
	done := false
	actions := make([]int, len(prices))
	portfolioSeries := make([]float64, len(prices))
	actionData := make([]plot.ActionData, len(prices))
	initialValue := marketEnv.PortfolioValue()

	for i := range actions {
		actions[i] = -1
		portfolioSeries[i] = marketEnv.PortfolioValue()
		actionData[i] = plot.ActionData{
			ActionName:   "nothing",
			AmountBought: 0.0,
			AmountSold:   0.0,
			Cash:         marketEnv.Cash(),
			Shares:       marketEnv.Shares(),
			Commission:   0.0,
		}
	}

	step := 0
	for !done {
		action := testAgent.Act(s)
		currentPrice := marketEnv.CurrentPrice()
		currentCash := marketEnv.Cash()
		currentShares := marketEnv.Shares()
		commission := marketEnv.Commission()

		// Calculate buy/sell amounts and commission before executing the action
		amountBought, amountSold, commissionPaid := calculateActionAmountsAndCommission(action, currentCash, currentShares, currentPrice, commission)

		next, _, d := marketEnv.Step(action)
		actions[step] = int(action)
		portfolioSeries[step+1] = marketEnv.PortfolioValue()

		// Get cash and shares after the action
		afterCash := marketEnv.Cash()
		afterShares := marketEnv.Shares()

		// Store action data at step+1 to match portfolioSeries indexing
		// (step 0 is initial state, step+1 is after first action)
		actionData[step+1] = plot.ActionData{
			ActionName:   action.String(),
			AmountBought: amountBought,
			AmountSold:   amountSold,
			Cash:         afterCash,
			Shares:       afterShares,
			Commission:   commissionPaid,
		}
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

	return portfolioSeries, actions, actionData
}

// testAgent is a simple agent that only acts (for testing).
type testAgent struct {
	policy agent.Actor
}

func (a *testAgent) Act(s state.State) agent.Action {
	return a.policy.Act(s)
}

// calculateActionAmountsAndCommission calculates the amount of shares bought or sold and commission paid for a given action.
func calculateActionAmountsAndCommission(action agent.Action, cash, shares, price, commission float64) (amountBought, amountSold, commissionPaid float64) {
	switch action {
	case agent.ActionBuySmall:
		cost := cash * agent.BuySmall
		commissionPaid = cost * commission
		amountBought = (cost - commissionPaid) / price
		amountSold = 0.0
	case agent.ActionBuyLarge:
		cost := cash * agent.BuyLarge
		commissionPaid = cost * commission
		amountBought = (cost - commissionPaid) / price
		amountSold = 0.0
	case agent.ActionSellSmall:
		if shares <= 0 {
			return 0.0, 0.0, 0.0
		}
		amountBought = 0.0
		amountSold = shares * agent.SellSmall
		proceeds := amountSold * price
		commissionPaid = proceeds * commission
	case agent.ActionSellLarge:
		if shares <= 0 {
			return 0.0, 0.0, 0.0
		}
		amountBought = 0.0
		amountSold = shares * agent.SellLarge
		proceeds := amountSold * price
		commissionPaid = proceeds * commission
	default:
		amountBought = 0.0
		amountSold = 0.0
		commissionPaid = 0.0
	}
	return amountBought, amountSold, commissionPaid
}

// loadTestPricesFromCSV loads price data from test.csv file.
// The CSV has columns: MSFT, IBM, SBUX, AAPL, GSPC, Date
// We'll use GSPC (S&P 500 index) column (index 4) as the price series.
func loadTestPricesFromCSV(filename string) ([]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header and one data row")
	}

	// GSPC is in column index 4 (5th column)
	gspcColIdx := 4
	if len(records[0]) <= gspcColIdx {
		return nil, fmt.Errorf("CSV file must have at least 5 columns")
	}

	// Skip header row, start from index 1
	prices := make([]float64, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		if len(records[i]) <= gspcColIdx {
			continue // Skip rows with insufficient columns
		}

		// Get GSPC price (index 4)
		priceStr := records[i][gspcColIdx]
		// Remove commas and quotes from the price string
		priceStr = strings.ReplaceAll(priceStr, ",", "")
		priceStr = strings.Trim(priceStr, `"`)
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse price at row %d: %w", i+1, err)
		}
		if price > 0 {
			prices = append(prices, price)
		}
	}

	return prices, nil
}
