package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kasaderos/rLportfolio/pkg/agent"
	"github.com/kasaderos/rLportfolio/pkg/env"
	"github.com/kasaderos/rLportfolio/pkg/plot"
	"github.com/kasaderos/rLportfolio/pkg/state"
	"github.com/kasaderos/rLportfolio/pkg/trainer"
)

const (
	// Q-learning parameters
	alpha   = 0.1
	gamma   = 0.95
	epsilon = 0.1

	episodes  = 1000
	minPrices = 50 // Minimum prices needed to start training

	// Local approximation parameters
	approxM = 7
	approxN = 5
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

	// Load all stock data from train.csv
	stockData, err := loadAllStocksFromCSV("data/train.csv")
	if err != nil {
		fmt.Printf("Error loading stocks from CSV: %v\n", err)
		return
	}

	if len(stockData) == 0 {
		fmt.Printf("Error: No stock data found\n")
		return
	}

	fmt.Printf("Loaded %d stocks from train.csv\n", len(stockData))
	for name, prices := range stockData {
		fmt.Printf("  %s: %d prices\n", name, len(prices))
	}

	// Create Q-table and policy (shared across all stocks)
	Q := agent.NewQTable(state.NumStates, agent.NumActions)
	policy := agent.NewEpsilonGreedyPolicy(Q.Q, epsilon, rng)

	// Create agent
	rlAgent := agent.NewQLearningAgent(Q, policy, alpha, gamma)

	// Train on each stock sequentially
	episodesPerStock := *episodeCount / len(stockData)
	if episodesPerStock < 1 {
		episodesPerStock = 1
	}

	fmt.Printf("\n=== Training on %d stocks ===\n", len(stockData))
	fmt.Printf("Episodes per stock: %d\n\n", episodesPerStock)

	stockNames := make([]string, 0, len(stockData))
	for name := range stockData {
		stockNames = append(stockNames, name)
	}
	// Sort for consistent ordering
	sort.Strings(stockNames)

	for _, stockName := range stockNames {
		prices := stockData[stockName]
		if len(prices) < minPrices {
			fmt.Printf("Skipping %s: Need at least %d prices, got %d\n", stockName, minPrices, len(prices))
			continue
		}

		fmt.Printf("Training on %s (%d prices)...\n", stockName, len(prices))

		// Create environment for this stock
		marketEnv := env.NewMarketEnv(env.MarketConfig{
			Prices:      prices,
			InitialCash: 10000.0,
			ApproxM:     approxM,
			ApproxN:     approxN,
			MinStartIdx: 120, // Need at least 120 for MA120
			Commission:  0.002,
		})

		// Create trainer
		t := trainer.NewTrainer(marketEnv, rlAgent)

		// Train on this stock
		t.Run(episodesPerStock, 100)
		fmt.Printf("Completed training on %s\n\n", stockName)
	}

	// Test the learned policy on the last stock (or first stock if available)
	var testPrices []float64
	var testStockName string
	if len(stockNames) > 0 {
		testStockName = stockNames[len(stockNames)-1]
		testPrices = stockData[testStockName]
	} else {
		// Fallback: use first available stock
		for name, prices := range stockData {
			testStockName = name
			testPrices = prices
			break
		}
	}

	if len(testPrices) >= minPrices {
		fmt.Printf("\n=== Testing Learned Policy on %s ===\n", testStockName)
		marketEnv := env.NewMarketEnv(env.MarketConfig{
			Prices:      testPrices,
			InitialCash: 10000.0,
			ApproxM:     approxM,
			ApproxN:     approxN,
			MinStartIdx: 120, // Need at least 120 for MA120
			Commission:  0.002,
		})

		portfolioSeries, actions, actionData := testPolicy(Q.Q, testPrices, marketEnv)

		// Save series data to data/series.csv
		if err := plot.SaveSeriesData(testPrices, portfolioSeries, actions, actionData); err != nil {
			fmt.Printf("Failed to save series: %v\n", err)
		} else {
			fmt.Println("Saved series data to data/series.csv")
		}
	}

	// Save Q-matrix to data/q_matrix.csv
	if err := plot.SaveQMatrixData(Q.Q); err != nil {
		fmt.Printf("Failed to save Q matrix: %v\n", err)
	} else {
		fmt.Println("Saved Q matrix to data/q_matrix.csv")
	}
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

// testAgent is a simple agent that only acts (for testing).
type testAgent struct {
	policy agent.Actor
}

func (a *testAgent) Act(s state.State) agent.Action {
	return a.policy.Act(s)
}

// loadAllStocksFromCSV loads all stock price data from a CSV file.
// Returns a map where keys are stock names and values are price arrays.
// The CSV should have a header row with stock names (excluding Date column).
func loadAllStocksFromCSV(filename string) (map[string][]float64, error) {
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

	// Parse header to find stock columns (exclude Date column)
	header := records[0]
	stockIndices := make(map[string]int)

	for i, colName := range header {
		colName = strings.Trim(colName, `"`)
		// Skip Date column
		if strings.ToLower(colName) == "date" {
			continue
		}
		stockIndices[colName] = i
	}

	if len(stockIndices) == 0 {
		return nil, fmt.Errorf("no stock columns found in CSV header")
	}

	// Initialize price arrays for each stock
	stockData := make(map[string][]float64)
	for stockName := range stockIndices {
		stockData[stockName] = make([]float64, 0, len(records)-1)
	}

	// Parse data rows
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) == 0 {
			continue
		}

		for stockName, colIdx := range stockIndices {
			if colIdx >= len(row) {
				continue
			}

			priceStr := row[colIdx]
			// Remove commas and quotes from the price string
			priceStr = strings.ReplaceAll(priceStr, ",", "")
			priceStr = strings.Trim(priceStr, `"`)
			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				// Skip invalid prices for this row/stock
				continue
			}
			if price > 0 {
				stockData[stockName] = append(stockData[stockName], price)
			}
		}
	}

	return stockData, nil
}
