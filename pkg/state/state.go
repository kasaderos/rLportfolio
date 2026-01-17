package state

// State represents the market state for portfolio optimization.
// It encodes market regime indicators from local approximation predictions.
type State struct {
	// Encoded state index (0 to numStates-1)
	Index int

	// Raw components (for debugging/analysis)
	ExpRetCat1    int // Expected return category from first approximation
	MinDistCat1   int // Minimum distance category from first approximation
	ExpRetCat2    int // Expected return category from second approximation
	MinDistCat2   int // Minimum distance category from second approximation
}

// Constants for state encoding
const (
	// Expected return categories
	ExpRetNeutral = iota
	ExpRetSmallPos
	ExpRetMedPos
	ExpRetLargePos
	ExpRetSmallNeg
	ExpRetMedNeg
	ExpRetLargeNeg
	NumExpRetCategories
)

const (
	// Minimum distance categories
	MinDistSmall = iota
	MinDistLarge
	NumMinDistCategories
)

const (
	// State space: (7 exp_ret × 2 min_distance)^2 = 196 states
	NumStatePerCall = NumExpRetCategories * NumMinDistCategories
	NumStates       = NumStatePerCall * NumStatePerCall
)

// NewState creates a new State from component categories.
func NewState(expRet1, minDist1, expRet2, minDist2 int) State {
	return State{
		Index:       Encode(expRet1, minDist1, expRet2, minDist2),
		ExpRetCat1:  expRet1,
		MinDistCat1: minDist1,
		ExpRetCat2:  expRet2,
		MinDistCat2: minDist2,
	}
}

// Encode encodes (exp_ret1, min_dist1, exp_ret2, min_dist2) into a single state index.
func Encode(expRet1, minDist1, expRet2, minDist2 int) int {
	return (expRet1*NumMinDistCategories+minDist1)*NumStatePerCall + (expRet2*NumMinDistCategories + minDist2)
}

// GetExpRetCategory maps expected return to category (0-6).
func GetExpRetCategory(expRet float64) int {
	if expRet < -0.02 {
		return ExpRetLargeNeg
	} else if expRet < -0.01 {
		return ExpRetMedNeg
	} else if expRet < -0.005 {
		return ExpRetSmallNeg
	} else if expRet > 0.02 {
		return ExpRetLargePos
	} else if expRet > 0.01 {
		return ExpRetMedPos
	} else if expRet > 0.005 {
		return ExpRetSmallPos
	}
	return ExpRetNeutral // -0.005 < exp_ret < 0.005
}

// GetMinDistCategory maps minimum distance to category.
func GetMinDistCategory(minDist float64) int {
	if minDist < 0.005 {
		return MinDistSmall
	}
	return MinDistLarge
}
