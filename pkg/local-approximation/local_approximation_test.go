package localapproximation

import (
	"math"
	"testing"
)

func TestLocalApproximation(t *testing.T) {
	tests := []struct {
		name      string
		returns   []float64
		m         int
		n         int
		wantError bool
		checkFunc func(t *testing.T, prediction float64, minDistance float64)
	}{
		{
			name:    "simple pattern matching",
			returns: []float64{1.0, 2.0, 3.0, 1.0, 2.0, 3.0, 1.0, 2.0},
			m:       3,
			n:       2,
			checkFunc: func(t *testing.T, prediction float64, minDistance float64) {
				// Last window: [1.0, 2.0, 3.0]
				// Should find exact match at positions 4 (window ending at index 4)
				// Next value after that window is 3.0
				// First distance should be 0 (exact match)
				if minDistance > 1e-10 {
					t.Errorf("expected min distance to be 0 (exact match), got %f", minDistance)
				}
				// Prediction should be close to 3.0
				if math.Abs(prediction-3.0) > 0.1 {
					t.Errorf("expected prediction close to 3.0, got %f", prediction)
				}
			},
		},
		{
			name:    "increasing trend",
			returns: []float64{10.0, 11.0, 12.0, 13.0, 14.0, 15.0, 16.0, 17.0, 18.0, 19.0},
			m:       3,
			n:       3,
			checkFunc: func(t *testing.T, prediction float64, minDistance float64) {
				// Last window: [17.0, 18.0, 19.0]
				// Prediction should be reasonable (within data range)
				if prediction < 10.0 || prediction > 25.0 {
					t.Errorf("expected prediction between 10.0 and 25.0, got %f", prediction)
				}
				// Check that min distance is valid
				if minDistance < 0 || math.IsNaN(minDistance) || math.IsInf(minDistance, 0) {
					t.Errorf("invalid min distance: %f", minDistance)
				}
			},
		},
		{
			name:    "single neighbor",
			returns: []float64{5.0, 6.0, 7.0, 5.5, 6.5, 7.5},
			m:       2,
			n:       1,
			checkFunc: func(t *testing.T, prediction float64, minDistance float64) {
				// Should have a valid prediction
				if math.IsNaN(prediction) || math.IsInf(prediction, 0) {
					t.Errorf("prediction is NaN or Inf: %f", prediction)
				}
				if minDistance < 0 || math.IsNaN(minDistance) || math.IsInf(minDistance, 0) {
					t.Errorf("invalid min distance: %f", minDistance)
				}
			},
		},
		{
			name:    "prediction multiple steps ahead",
			returns: []float64{1.0, 2.0, 3.0, 4.0, 1.0, 2.0, 3.0, 4.0, 5.0, 1.0, 2.0, 3.0},
			m:       3,
			n:       2,
			checkFunc: func(t *testing.T, prediction float64, minDistance float64) {
				// Last window: [1.0, 2.0, 3.0]
				// Should find pattern at position 3 (window ending at index 3), next 2 steps is 5.0
				// Prediction should be reasonable
				if math.IsNaN(prediction) || math.IsInf(prediction, 0) {
					t.Errorf("prediction is NaN or Inf: %f", prediction)
				}
				if minDistance < 0 || math.IsNaN(minDistance) || math.IsInf(minDistance, 0) {
					t.Errorf("invalid min distance: %f", minDistance)
				}
			},
		},
		{
			name:      "empty prices",
			returns:   []float64{},
			m:         3,
			n:         2,
			wantError: true,
		},
		{
			name:      "invalid m (zero)",
			returns:   []float64{1.0, 2.0, 3.0},
			m:         0,
			n:         2,
			wantError: true,
		},
		{
			name:      "invalid m (negative)",
			returns:   []float64{1.0, 2.0, 3.0},
			m:         -1,
			n:         2,
			wantError: true,
		},
		{
			name:      "invalid n (zero)",
			returns:   []float64{1.0, 2.0, 3.0},
			m:         2,
			n:         0,
			wantError: true,
		},
		{
			name:      "insufficient data (need m+p)",
			returns:   []float64{1.0, 2.0},
			m:         2,
			n:         1,
			wantError: true,
		},
		{
			name:      "insufficient historical data",
			returns:   []float64{1.0, 2.0, 3.0},
			m:         2,
			n:         1,
			wantError: true,
		},
		{
			name:    "n greater than available neighbors",
			returns: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			m:       2,
			n:       100, // More than available
			checkFunc: func(t *testing.T, prediction float64, minDistance float64) {
				if minDistance < 0 || math.IsNaN(minDistance) || math.IsInf(minDistance, 0) {
					t.Errorf("invalid min distance: %f", minDistance)
				}
			},
		},
		{
			name:    "exclude last window from nearest search",
			returns: []float64{5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 1.0, 2.0, 3.0},
			m:       3,
			n:       2,
			checkFunc: func(t *testing.T, prediction float64, minDistance float64) {
				// Last window: [1.0, 2.0, 3.0]
				// No other identical window exists in earlier history.
				// If the last window were (incorrectly) included, minDistance would be 0.
				if minDistance <= 1e-10 {
					t.Errorf("expected min distance > 0 (exclude last window), got %f", minDistance)
				}
				// Prediction should still be valid
				if math.IsNaN(prediction) || math.IsInf(prediction, 0) {
					t.Errorf("prediction is NaN or Inf: %f", prediction)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prediction, minDistance, err := LocalApproximation(tt.returns, tt.m, tt.n)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, prediction, minDistance)
			}
		})
	}
}

func TestLocalApproximation_RealisticPriceData(t *testing.T) {
	// Simulate realistic price data with some patterns
	returns := []float64{
		100.0, 101.0, 102.0, 101.0, 100.0, // Pattern 1
		110.0, 111.0, 112.0, 111.0, 110.0, // Pattern 2
		105.0, 106.0, 107.0, 106.0, 105.0, // Pattern 3 (similar to pattern 1)
		100.0, 101.0, 102.0, // Last window (matches pattern 1 start)
	}

	m := 3
	n := 3

	prediction, minDistance, err := LocalApproximation(returns, m, n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that prediction is reasonable (not NaN or Inf)
	if math.IsNaN(prediction) || math.IsInf(prediction, 0) {
		t.Errorf("prediction is NaN or Inf: %f", prediction)
	}

	// Check that all distances are non-negative
	if minDistance < 0 {
		t.Errorf("min distance is negative: %f", minDistance)
	}

	// Last window is [100.0, 101.0, 102.0]
	// Should find similar patterns and predict around 101.0 or 100.0 (following the pattern)
	if prediction < 90.0 || prediction > 115.0 {
		t.Errorf("prediction seems unreasonable: %f (expected between 90.0 and 115.0)", prediction)
	}
}

func TestLocalApproximation_ExactMatch(t *testing.T) {
	// Test case where we have an exact match
	returns := []float64{
		5.0, 10.0, 15.0, // Pattern
		20.0, 25.0, 30.0, // Different values
		5.0, 10.0, 15.0, // Exact match of first pattern
	}

	m := 3
	n := 1

	_, minDistance, err := LocalApproximation(returns, m, n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find exact match (distance = 0)
	// The exact match should have distance 0 (or very close to 0)
	if minDistance > 1e-10 {
		t.Errorf("expected distance 0 for exact match, got %f", minDistance)
	}
}
