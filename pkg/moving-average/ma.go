package movingaverage

import (
	"math"
	"sort"
)

// MAPeriods defines the moving average periods to use.
var MAPeriods = []int{5, 10, 20, 40, 80, 120}

const (
	// MA5, MA10, MA20, MA40, MA80, MA120 represent the moving average indices
	MA5   = 1
	MA10  = 2
	MA20  = 3
	MA40  = 4
	MA80  = 5
	MA120 = 6
	Price = 7 // Price is represented as 7 in the ordering
)

// periodToIndex maps period to index (1-6)
var periodToIndex = map[int]int{
	5:   MA5,
	10:  MA10,
	20:  MA20,
	40:  MA40,
	80:  MA80,
	120: MA120,
}

// CalculateMA calculates a simple moving average for the given period.
func CalculateMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}

	ma := make([]float64, len(prices)-period+1)
	for i := period - 1; i < len(prices); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += prices[j]
		}
		ma[i-period+1] = sum / float64(period)
	}

	return ma
}

// CalculateAllMAs calculates all moving averages (5, 10, 20, 40, 80, 120) for the given prices.
// Returns a map where keys are periods and values are MA arrays.
// Note: Each MA array will have different lengths (shorter for longer periods).
func CalculateAllMAs(prices []float64) map[int][]float64 {
	mas := make(map[int][]float64)
	for _, period := range MAPeriods {
		mas[period] = CalculateMA(prices, period)
	}
	return mas
}

// ValueWithIndex represents a value with its identifier (MA period or Price).
type ValueWithIndex struct {
	Value float64
	Index int // 1-6 for MAs, 7 for Price
}

// GetMAOrdering determines the vertical ordering of moving averages and current price.
// Returns a slice representing the order from top (highest) to bottom (lowest).
// Values: 1=MA5, 2=MA10, 3=MA20, 4=MA40, 5=MA80, 6=MA120, 7=Price
// Always returns exactly 7 elements. Assumes idx >= 120 (all MAs available).
func GetMAOrdering(prices []float64, idx int) []int {
	if idx < 0 || idx >= len(prices) {
		return nil
	}

	currentPrice := prices[idx]

	// Pre-allocate with exact size
	values := make([]ValueWithIndex, 7)

	// Calculate only the last MA value for each period (more efficient than calculating all MAs)
	// Assumes idx >= 120, so all periods have enough data
	for i, period := range MAPeriods {
		// Calculate MA value directly: sum of last 'period' prices
		sum := 0.0
		start := idx - period + 1
		for j := start; j <= idx; j++ {
			sum += prices[j]
		}
		maValue := sum / float64(period)

		// Map period to index using the mapping
		values[i] = ValueWithIndex{
			Value: maValue,
			Index: periodToIndex[period],
		}
	}

	// Add current price
	values[6] = ValueWithIndex{
		Value: currentPrice,
		Index: Price,
	}

	// Sort by value (descending - highest first)
	// For 7 elements, this is very fast
	sort.Slice(values, func(i, j int) bool {
		diff := values[i].Value - values[j].Value
		if math.Abs(diff) < 1e-10 {
			// If values are equal, maintain original order (by index)
			return values[i].Index < values[j].Index
		}
		return diff > 0
	})

	// Extract ordering
	ordering := make([]int, 7)
	for i := range values {
		ordering[i] = values[i].Index
	}

	return ordering
}

// EncodeMAState encodes the MA ordering into a state index.
// The ordering is a permutation of [1,2,3,4,5,6,7] representing MA5, MA10, MA20, MA40, MA80, MA120, Price.
// Returns a unique integer state index (0 to 5039, since 7! = 5040).
func EncodeMAState(ordering []int) int {
	if len(ordering) != 7 {
		return 0
	}

	// Use factorial number system (Lehmer code) to encode permutation
	// This gives us a unique index for each permutation
	state := 0
	factorials := []int{720, 120, 24, 6, 2, 1, 1} // 6!, 5!, 4!, 3!, 2!, 1!, 0!
	used := make([]bool, 8)                       // 1-indexed, so we need 8 elements (indices 0-7, use 1-7)

	for i := 0; i < 7; i++ {
		// Count how many unused numbers are smaller than ordering[i]
		count := 0
		for j := 1; j < ordering[i]; j++ {
			if !used[j] {
				count++
			}
		}
		state += count * factorials[i]
		used[ordering[i]] = true
	}

	return state
}

// DecodeMAState decodes a state index back into an ordering.
func DecodeMAState(stateIndex int) []int {
	ordering := make([]int, 7)
	used := make([]bool, 8)
	factorials := []int{720, 120, 24, 6, 2, 1, 1} // 6!, 5!, 4!, 3!, 2!, 1!, 0!

	for i := 0; i < 7; i++ {
		fact := factorials[i]
		pos := stateIndex / fact
		stateIndex %= fact

		count := 0
		for j := 1; j <= 7; j++ {
			if !used[j] {
				if count == pos {
					ordering[i] = j
					used[j] = true
					break
				}
				count++
			}
		}
	}

	return ordering
}

// GetMAStateForIndex calculates the MA ordering state for a given price index.
func GetMAStateForIndex(prices []float64, idx int) int {
	ordering := GetMAOrdering(prices, idx)
	if ordering == nil || len(ordering) != 7 {
		return 0
	}
	return EncodeMAState(ordering)
}

// NumMAStates returns the total number of possible MA ordering states.
// This is 7! = 5040 (all permutations of 7 elements).
func NumMAStates() int {
	return 5040
}

// GetMADivergenceState determines if moving averages are converging or diverging.
// Returns: 0 = converging, 1 = neutral, 2 = diverging
// Compares the current spread of MAs to the spread at a previous point.
func GetMADivergenceState(prices []float64, idx int) int {
	if idx < 120 || idx < 10 {
		return 1 // Neutral if not enough data
	}

	// Calculate current MA values
	currentMAs := make([]float64, len(MAPeriods))
	for i, period := range MAPeriods {
		sum := 0.0
		start := idx - period + 1
		for j := start; j <= idx; j++ {
			sum += prices[j]
		}
		currentMAs[i] = sum / float64(period)
	}

	// Calculate current spread (range: max - min)
	currentMax := currentMAs[0]
	currentMin := currentMAs[0]
	for _, ma := range currentMAs {
		if ma > currentMax {
			currentMax = ma
		}
		if ma < currentMin {
			currentMin = ma
		}
	}
	currentSpread := currentMax - currentMin

	// Calculate previous spread (10 periods ago, but ensure we have enough data)
	prevIdx := idx - 10
	if prevIdx < 120 {
		// If we can't go back 10 periods, use the earliest valid point (120)
		// In this case, we'll return neutral since we can't make a comparison
		if idx == 120 {
			return 1 // Neutral - can't compare yet
		}
		prevIdx = 120
	}

	prevMAs := make([]float64, len(MAPeriods))
	for i, period := range MAPeriods {
		sum := 0.0
		start := prevIdx - period + 1
		if start < 0 {
			start = 0
		}
		for j := start; j <= prevIdx; j++ {
			sum += prices[j]
		}
		prevMAs[i] = sum / float64(period)
	}

	prevMax := prevMAs[0]
	prevMin := prevMAs[0]
	for _, ma := range prevMAs {
		if ma > prevMax {
			prevMax = ma
		}
		if ma < prevMin {
			prevMin = ma
		}
	}
	prevSpread := prevMax - prevMin

	// Determine convergence/divergence
	// Use a threshold to avoid noise (1% of average price)
	avgPrice := (currentMax + currentMin) / 2.0
	threshold := avgPrice * 0.01

	spreadChange := currentSpread - prevSpread
	if spreadChange < -threshold {
		return 0 // Converging (spread decreased)
	} else if spreadChange > threshold {
		return 2 // Diverging (spread increased)
	}
	return 1 // Neutral (spread stable)
}
