package localapproximation

import (
	"errors"
	"math"
	"slices"
	"sort"
)

// Error definitions
var (
	ErrEmptyPrices                = errors.New("returns slice is empty")
	ErrInvalidWindowSize          = errors.New("window size m must be at least 1")
	ErrInvalidNeighborsCount      = errors.New("number of neighbors n must be at least 1")
	ErrInsufficientData           = errors.New("insufficient data: need at least m+p returns")
	ErrInsufficientHistoricalData = errors.New("insufficient historical data to find patterns")
	ErrNoPatternsFound            = errors.New("no historical patterns found")
)

// LocalApproximation implements the k-Nearest Neighbors (k-NN) Local Approximation Method (LAM).
// It finds similar historical patterns using L2 distance and uses their next values for prediction.
//
// Parameters:
//   - returns: slice of return values [a₁, a₂, ..., aₙ]
//   - m: size of the window (number of last returns to use as target pattern)
//   - n: number of nearest neighbors to use for prediction
//
// Returns:
//   - prediction: forecasted value 1 step ahead
//   - distances: L2 distances for the n nearest neighbors (sorted, smallest first)
//   - error: if insufficient data or invalid parameters
func LocalApproximation(returns []float64, m int, n int) (float64, float64, error) {
	p := 1
	inf := 1000.0

	if len(returns) == 0 {
		return 0, inf, ErrEmptyPrices
	}
	if m < 1 {
		return 0, inf, ErrInvalidWindowSize
	}

	if n < 1 {
		return 0, inf, ErrInvalidNeighborsCount
	}
	if len(returns) < m+p {
		return 0, inf, ErrInsufficientData
	}

	// Extract target window: last m returns [aₙ₋ₘ₊₁, ..., aₙ]
	targetWindow := returns[len(returns)-m:]

	// Minimum position to start searching (need m elements before + p elements after)
	minStart := m - 1
	// Do not allow the target (last) window as a candidate.
	// Window end index i must allow a next value at i+p, and i must be < len(returns)-1.
	maxStart := len(returns) - 1 - p

	if maxStart < minStart {
		return 0, inf, ErrInsufficientHistoricalData
	}

	// Type to store distance and position
	type neighbor struct {
		distance float64
		position int // position where the window ends (i in the algorithm)
	}

	var neighbors []neighbor

	// Slide through historical data to find similar windows
	for i := minStart; i <= maxStart; i++ {
		// Extract historical window: [aᵢ₋ₘ₊₁, ..., aᵢ]
		windowStart := i - m + 1
		historicalWindow := returns[windowStart : windowStart+m]

		// Calculate L2 distance: d = √(Σⱼ (target[j] - historical[j])²)
		var sumSquaredDiff float64
		for j := 0; j < m; j++ {
			diff := targetWindow[j] - historicalWindow[j]
			sumSquaredDiff += diff * diff
		}
		distance := math.Sqrt(sumSquaredDiff)

		neighbors = append(neighbors, neighbor{
			distance: distance,
			position: i,
		})
	}

	if len(neighbors) == 0 {
		return 0, inf, ErrNoPatternsFound
	}

	// Sort by distance (smallest first)
	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].distance < neighbors[j].distance
	})

	// Select n nearest neighbors
	numNeighbors := n
	if numNeighbors > len(neighbors) {
		numNeighbors = len(neighbors)
	}

	selectedNeighbors := neighbors[:numNeighbors]

	// Extract distances for return
	distances := make([]float64, numNeighbors)
	nextValues := make([]float64, numNeighbors)

	// Get next p-th element after each neighbor window
	for i, neighbor := range selectedNeighbors {
		distances[i] = neighbor.distance
		// Position i ends the window, so next p-th element is at position i + p
		nextValues[i] = returns[neighbor.position+p]
	}

	// Calculate weighted average prediction
	// Using inverse distance weighting: w_i = 1 / (d_i + ε)
	epsilon := 1e-10 // small value to avoid division by zero
	var weightedSum, weightSum float64

	for i := 0; i < numNeighbors; i++ {
		weight := 1.0 / (distances[i] + epsilon)
		weightedSum += weight * nextValues[i]
		weightSum += weight
	}

	prediction := weightedSum / weightSum

	return prediction, slices.Min(distances), nil
}
