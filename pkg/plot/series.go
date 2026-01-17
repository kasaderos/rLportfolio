package plot

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// SaveSeries saves price, cash, and action series to CSV files in plot/series directory.
func SaveSeries(prices []float64, cashSeries []float64, actions []int) error {
	// Create plot/series directory if it doesn't exist
	dir := "plot/series"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Save prices
	if err := saveToCSV(filepath.Join(dir, "prices.csv"), []string{"time", "price"}, prices); err != nil {
		return fmt.Errorf("failed to save prices: %w", err)
	}

	// Save cash series
	if err := saveToCSV(filepath.Join(dir, "cash.csv"), []string{"time", "cash"}, cashSeries); err != nil {
		return fmt.Errorf("failed to save cash series: %w", err)
	}

	// Save actions
	if err := saveActionsToCSV(filepath.Join(dir, "actions.csv"), []string{"time", "action"}, actions); err != nil {
		return fmt.Errorf("failed to save actions: %w", err)
	}

	return nil
}

// saveToCSV saves a single float64 series to CSV with header.
func saveToCSV(filepath string, header []string, data []float64) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for i, v := range data {
		record := []string{
			strconv.Itoa(i),
			strconv.FormatFloat(v, 'f', 6, 64),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return writer.Error()
}

// saveActionsToCSV saves actions to CSV.
func saveActionsToCSV(filepath string, header []string, actions []int) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for i, a := range actions {
		record := []string{
			strconv.Itoa(i),
			strconv.Itoa(a),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return writer.Error()
}
