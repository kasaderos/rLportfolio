package plot

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// ActionData represents action information for saving to CSV.
type ActionData struct {
	ActionName   string
	AmountBought float64
	AmountSold   float64
	Cash         float64
	Shares       float64
	Commission   float64
}

// SaveSeriesData saves all series data to a CSV file in data directory.
// The portfolioSeries should contain portfolio values (cash + price * shares) for each time step.
func SaveSeriesData(prices []float64, portfolioSeries []float64, actions []int, actionData []ActionData) error {
	return SaveSeriesDataToFile(prices, portfolioSeries, actions, actionData, "data/series.csv")
}

// SaveSeriesDataToFile saves all series data to a specified CSV file.
// The portfolioSeries should contain portfolio values (cash + price * shares) for each time step.
func SaveSeriesDataToFile(prices []float64, portfolioSeries []float64, actions []int, actionData []ActionData, filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"time", "price", "portfolio_value", "action", "action_name", "amount_bought", "amount_sold", "cash", "shares", "commission"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data - ensure all arrays have the same length
	maxLen := len(prices)
	if len(portfolioSeries) > maxLen {
		maxLen = len(portfolioSeries)
	}
	if len(actions) > maxLen {
		maxLen = len(actions)
	}
	if len(actionData) > maxLen {
		maxLen = len(actionData)
	}

	for i := 0; i < maxLen; i++ {
		price := 0.0
		if i < len(prices) {
			price = prices[i]
		}
		portfolioValue := 0.0
		if i < len(portfolioSeries) {
			portfolioValue = portfolioSeries[i]
		}
		action := -1
		if i < len(actions) {
			action = actions[i]
		}

		actionName := ""
		amountBought := 0.0
		amountSold := 0.0
		cash := 0.0
		shares := 0.0
		commission := 0.0
		if i < len(actionData) {
			actionName = actionData[i].ActionName
			amountBought = actionData[i].AmountBought
			amountSold = actionData[i].AmountSold
			cash = actionData[i].Cash
			shares = actionData[i].Shares
			commission = actionData[i].Commission
		}

		record := []string{
			strconv.Itoa(i),
			strconv.FormatFloat(price, 'f', 6, 64),
			strconv.FormatFloat(portfolioValue, 'f', 6, 64),
			strconv.Itoa(action),
			actionName,
			strconv.FormatFloat(amountBought, 'f', 6, 64),
			strconv.FormatFloat(amountSold, 'f', 6, 64),
			strconv.FormatFloat(cash, 'f', 6, 64),
			strconv.FormatFloat(shares, 'f', 6, 64),
			strconv.FormatFloat(commission, 'f', 6, 64),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return writer.Error()
}

// SaveQMatrixData saves the Q-matrix to CSV in data directory.
func SaveQMatrixData(Q [][]float64) error {
	// Create data directory if it doesn't exist
	dir := "data"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	filepath := filepath.Join(dir, "q_matrix.csv")
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header: action indices
	header := make([]string, len(Q[0])+1)
	header[0] = "state"
	for i := 0; i < len(Q[0]); i++ {
		header[i+1] = "action_" + strconv.Itoa(i)
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write each state row
	for state, row := range Q {
		record := make([]string, len(row)+1)
		record[0] = strconv.Itoa(state)
		for i, v := range row {
			record[i+1] = strconv.FormatFloat(v, 'f', 6, 64)
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write row for state %d: %w", state, err)
		}
	}

	return writer.Error()
}

// LoadQMatrixData loads the Q-matrix from data/q_matrix.csv.
func LoadQMatrixData() ([][]float64, error) {
	filepath := "data/q_matrix.csv"
	file, err := os.Open(filepath)
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
		return nil, fmt.Errorf("insufficient data in file")
	}

	// First row is header, determine number of actions from header
	numActions := len(records[0]) - 1 // Subtract 1 for state column
	if numActions <= 0 {
		return nil, fmt.Errorf("invalid header format")
	}

	// Parse Q-matrix
	var Q [][]float64
	for i := 1; i < len(records); i++ {
		if len(records[i]) < numActions+1 {
			continue
		}

		row := make([]float64, numActions)
		for j := 0; j < numActions; j++ {
			val, err := strconv.ParseFloat(records[i][j+1], 64) // j+1 to skip state column
			if err != nil {
				return nil, fmt.Errorf("failed to parse value at row %d, col %d: %w", i+1, j+1, err)
			}
			row[j] = val
		}
		Q = append(Q, row)
	}

	return Q, nil
}

// LoadSeriesData loads series data from data/series.csv.
// Returns prices, portfolio values (cash + price * shares), and actions.
// The new columns (action_name, amount_bought, amount_sold) are ignored for backward compatibility.
func LoadSeriesData() ([]float64, []float64, []int, error) {
	filepath := "data/series.csv"
	file, err := os.Open(filepath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comment = '#'

	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, nil, nil, fmt.Errorf("insufficient data in file")
	}

	// Skip header
	var prices []float64
	var portfolioSeries []float64
	var actions []int

	for i := 1; i < len(records); i++ {
		if len(records[i]) < 4 {
			continue
		}

		price, err := strconv.ParseFloat(records[i][1], 64)
		if err != nil {
			continue
		}
		portfolioValue, err := strconv.ParseFloat(records[i][2], 64)
		if err != nil {
			continue
		}
		action, err := strconv.Atoi(records[i][3])
		if err != nil {
			continue
		}

		prices = append(prices, price)
		portfolioSeries = append(portfolioSeries, portfolioValue)
		actions = append(actions, action)
	}

	return prices, portfolioSeries, actions, nil
}
