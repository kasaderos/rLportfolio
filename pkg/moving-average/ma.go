package movingaverage

import (
	"math"
	"sort"
)

// MAPeriods defines the moving average periods to use.
var MAPeriods = []int{10, 20, 30, 40, 50, 60}

const (
	// MA10, MA20, MA30, MA40, MA50, MA60 represent the moving average indices
	MA10  = 1
	MA20  = 2
	MA30  = 3
	MA40  = 4
	MA50  = 5
	MA60  = 6
	Price = 7 // Price is represented as 7 in the ordering
)

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

// CalculateAllMAs calculates all moving averages (10, 20, 30, 40, 50, 60) for the given prices.
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
// Values: 1=MA10, 2=MA20, 3=MA30, 4=MA40, 5=MA50, 6=MA60, 7=Price
// Always returns exactly 7 elements. Assumes idx >= 60 (all MAs available).
func GetMAOrdering(prices []float64, idx int) []int {
	if idx < 0 || idx >= len(prices) {
		return nil
	}

	currentPrice := prices[idx]

	// Pre-allocate with exact size
	values := make([]ValueWithIndex, 7)

	// Calculate only the last MA value for each period (more efficient than calculating all MAs)
	// Assumes idx >= 60, so all periods have enough data
	for i, period := range MAPeriods {
		// Calculate MA value directly: sum of last 'period' prices
		sum := 0.0
		start := idx - period + 1
		for j := start; j <= idx; j++ {
			sum += prices[j]
		}
		maValue := sum / float64(period)

		// Map period to index: 10->1, 20->2, 30->3, 40->4, 50->5, 60->6
		values[i] = ValueWithIndex{
			Value: maValue,
			Index: period / 10,
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
// The ordering is a permutation of [1,2,3,4,5,6,7] representing MA10, MA20, MA30, MA40, MA50, MA60, Price.
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
