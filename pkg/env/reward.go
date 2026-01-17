package env

import "math"

// CalculateReward calculates the reward as log return of portfolio value.
func CalculateReward(portfolioValueBefore, portfolioValueAfter float64) float64 {
	if portfolioValueBefore > 0 {
		return math.Log(portfolioValueAfter / portfolioValueBefore)
	}
	return 0
}
