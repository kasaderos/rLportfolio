package plot

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// SaveQMatrix saves the Q-matrix to a CSV file in plot/matrix directory.
// Each row corresponds to a state, and columns are actions.
func SaveQMatrix(Q [][]float64) error {
	// Create plot/matrix directory if it doesn't exist
	dir := "plot/matrix"
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
