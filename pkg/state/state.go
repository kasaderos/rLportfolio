package state

// State represents the market state for portfolio optimization.
// It encodes market regime indicators from local approximation predictions and portfolio position.
type State struct {
	// Encoded state index (0 to numStates-1)
	Index int

	// Raw components (for debugging/analysis)
	ExpRetCat  int // Expected return category from approximation
	MinDistCat int // Minimum distance category from approximation
	CashCat    int // Cash position category
	SharesCat  int // Shares position category
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
	// Portfolio position categories (3 categories: None, Medium, High)
	PosNone = iota
	PosMedium
	PosHigh
	NumPositionCategories = 3
)

const (
	// Market state space: 7 exp_ret × 2 min_distance = 14 states
	NumMarketStates = NumExpRetCategories * NumMinDistCategories
	// Total state space: market states × cash categories × shares categories
	NumStates = NumMarketStates * NumPositionCategories * NumPositionCategories
)

// NewState creates a new State from component categories.
func NewState(expRet, minDist, cashCat, sharesCat int) State {
	return State{
		Index:      Encode(expRet, minDist, cashCat, sharesCat),
		ExpRetCat:  expRet,
		MinDistCat: minDist,
		CashCat:    cashCat,
		SharesCat:  sharesCat,
	}
}

// Encode encodes (exp_ret, min_dist, cash_cat, shares_cat) into a single state index.
func Encode(expRet, minDist, cashCat, sharesCat int) int {
	marketState := expRet*NumMinDistCategories + minDist
	return marketState*NumPositionCategories*NumPositionCategories + cashCat*NumPositionCategories + sharesCat
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

// GetCashCategory maps cash percentage of portfolio to category.
// portfolioValue is the total portfolio value (cash + shares * price).
func GetCashCategory(cash, portfolioValue float64) int {
	if portfolioValue <= 0 {
		return PosNone
	}
	cashRatio := cash / portfolioValue
	if cashRatio < 0.2 {
		return PosNone
	} else if cashRatio < 0.8 {
		return PosMedium
	}
	return PosHigh
}

// GetSharesCategory maps shares percentage of portfolio to category.
// portfolioValue is the total portfolio value (cash + shares * price).
func GetSharesCategory(sharesValue, portfolioValue float64) int {
	if portfolioValue <= 0 {
		return PosNone
	}
	sharesRatio := sharesValue / portfolioValue
	if sharesRatio < 0.2 {
		return PosNone
	} else if sharesRatio < 0.8 {
		return PosMedium
	}
	return PosHigh
}
