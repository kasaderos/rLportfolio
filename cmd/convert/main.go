package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run cmd/convert/main.go <input.csv> <output.csv>")
		fmt.Println("Example: go run cmd/convert/main.go data/tsla.csv data/test.csv")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// Read input CSV
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV: %v\n", err)
		os.Exit(1)
	}

	if len(records) < 2 {
		fmt.Println("Error: CSV must have at least a header and one data row")
		os.Exit(1)
	}

	// Find columns
	header := records[0]
	dateColIdx := -1
	priceColIdx := -1

	for i, col := range header {
		col = strings.ToLower(strings.Trim(col, `"`))
		if col == "date" {
			dateColIdx = i
		} else if col == "close/last" {
			priceColIdx = i
		}
	}

	if dateColIdx < 0 {
		fmt.Println("Error: Could not find Date column")
		os.Exit(1)
	}
	if priceColIdx < 0 {
		fmt.Println("Error: Could not find Close/Last column")
		os.Exit(1)
	}

	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Write header matching train.csv format: "MSFT","IBM","SBUX","AAPL","GSPC","Date"
	headerRow := []string{`TSLA`, `Date`}
	if err := writer.Write(headerRow); err != nil {
		fmt.Printf("Error writing header: %v\n", err)
		os.Exit(1)
	}

	// Process data rows (skip header, process in reverse to match chronological order)
	for i := len(records) - 1; i >= 1; i-- {
		row := records[i]
		if len(row) <= dateColIdx || len(row) <= priceColIdx {
			continue
		}

		// Parse date (MM/DD/YYYY format)
		dateStr := strings.Trim(row[dateColIdx], `"`)
		date, err := time.Parse("01/02/2006", dateStr)
		if err != nil {
			// Try alternative format
			date, err = time.Parse("1/2/2006", dateStr)
			if err != nil {
				fmt.Printf("Warning: Could not parse date at row %d: %s\n", i+1, dateStr)
				continue
			}
		}
		formattedDate := date.Format("2006-01-02")

		// Parse price (remove $ and commas)
		priceStr := strings.Trim(row[priceColIdx], `"`)
		priceStr = strings.ReplaceAll(priceStr, "$", "")
		priceStr = strings.ReplaceAll(priceStr, ",", "")
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			fmt.Printf("Warning: Could not parse price at row %d: %s\n", i+1, priceStr)
			continue
		}

		// Write row: TSLA price repeated 5 times (to match train.csv structure), then date
		// Format matches train.csv: unquoted numbers for prices, quoted date
		outputRow := []string{
			fmt.Sprintf("%.6f", price),
			fmt.Sprintf("%s", formattedDate),
		}
		if err := writer.Write(outputRow); err != nil {
			fmt.Printf("Error writing row: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Successfully converted %s to %s\n", inputFile, outputFile)
	fmt.Printf("Converted %d data rows\n", len(records)-1)
}
