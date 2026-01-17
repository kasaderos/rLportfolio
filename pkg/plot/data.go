package plot

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// SaveSeriesData saves all series data to a CSV file in rl/plot/data directory.
func SaveSeriesData(prices []float64, cashSeries []float64, actions []int) error {
	// Create rl/plot/data directory if it doesn't exist
	dir := "rl/plot/data"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	filepath := filepath.Join(dir, "series.csv")
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"time", "price", "cash", "action"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data - ensure all arrays have the same length
	maxLen := len(prices)
	if len(cashSeries) > maxLen {
		maxLen = len(cashSeries)
	}
	if len(actions) > maxLen {
		maxLen = len(actions)
	}

	for i := 0; i < maxLen; i++ {
		price := 0.0
		if i < len(prices) {
			price = prices[i]
		}
		cash := 0.0
		if i < len(cashSeries) {
			cash = cashSeries[i]
		}
		action := -1
		if i < len(actions) {
			action = actions[i]
		}

		record := []string{
			strconv.Itoa(i),
			strconv.FormatFloat(price, 'f', 6, 64),
			strconv.FormatFloat(cash, 'f', 6, 64),
			strconv.Itoa(action),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return writer.Error()
}

// SaveQMatrixData saves the Q-matrix to CSV in rl/plot/data directory.
func SaveQMatrixData(Q [][]float64) error {
	// Create rl/plot/data directory if it doesn't exist
	dir := "rl/plot/data"
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

// LoadSeriesData loads series data from rl/plot/data/series.csv.
func LoadSeriesData() ([]float64, []float64, []int, error) {
	filepath := "rl/plot/data/series.csv"
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
	var cashSeries []float64
	var actions []int

	for i := 1; i < len(records); i++ {
		if len(records[i]) < 4 {
			continue
		}

		price, err := strconv.ParseFloat(records[i][1], 64)
		if err != nil {
			continue
		}
		cash, err := strconv.ParseFloat(records[i][2], 64)
		if err != nil {
			continue
		}
		action, err := strconv.Atoi(records[i][3])
		if err != nil {
			continue
		}

		prices = append(prices, price)
		cashSeries = append(cashSeries, cash)
		actions = append(actions, action)
	}

	return prices, cashSeries, actions, nil
}
