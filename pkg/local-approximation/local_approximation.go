package localapproximation

import (
	"container/heap"
	"errors"
	"math"
	"slices"
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

// neighbor stores distance and position information for k-NN search
type neighbor struct {
	distance float64
	position int // position where the window ends (i in the algorithm)
}

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

	// Use a max-heap to keep only the n nearest neighbors
	// This avoids storing and sorting all neighbors, reducing memory and time complexity
	// Max-heap implementation (largest distance at root)
	neighborHeap := make(neighborMaxHeap, 0, n+1)
	heap.Init(&neighborHeap)

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

		// Add to heap if we have space or if this is closer than the farthest neighbor
		if neighborHeap.Len() < n {
			heap.Push(&neighborHeap, neighbor{
				distance: distance,
				position: i,
			})
		} else if distance < neighborHeap[0].distance {
			// Replace the farthest neighbor with this closer one
			neighborHeap[0] = neighbor{
				distance: distance,
				position: i,
			}
			heap.Fix(&neighborHeap, 0)
		}
	}

	if neighborHeap.Len() == 0 {
		return 0, inf, ErrNoPatternsFound
	}

	// Extract neighbors from heap and sort by distance (smallest first)
	// Since we only have n elements, this is O(n log n) instead of O(k log k)
	numNeighbors := neighborHeap.Len()
	distances := make([]float64, numNeighbors)
	nextValues := make([]float64, numNeighbors)

	// Pop all neighbors from heap and collect them
	neighbors := make([]neighbor, 0, numNeighbors)
	for neighborHeap.Len() > 0 {
		neighbors = append(neighbors, heap.Pop(&neighborHeap).(neighbor))
	}

	// Sort by distance (smallest first) - only n elements, so this is fast
	slices.SortFunc(neighbors, func(a, b neighbor) int {
		if a.distance < b.distance {
			return -1
		}
		if a.distance > b.distance {
			return 1
		}
		return 0
	})

	// Extract distances and next values
	for i, neighbor := range neighbors {
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

// neighborMaxHeap implements a max-heap for neighbors (largest distance at root)
type neighborMaxHeap []neighbor

func (h neighborMaxHeap) Len() int           { return len(h) }
func (h neighborMaxHeap) Less(i, j int) bool { return h[i].distance > h[j].distance } // Max-heap: larger distance is "less"
func (h neighborMaxHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *neighborMaxHeap) Push(x any) {
	*h = append(*h, x.(neighbor))
}

func (h *neighborMaxHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
