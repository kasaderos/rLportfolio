package state

// State represents the market state for portfolio optimization.
// It encodes moving average ordering state and portfolio position.
type State struct {
	// Encoded state index (0 to numStates-1)
	Index int

	// Raw components (for debugging/analysis)
	MAState   int // Moving average ordering state (0-5039)
	CashCat   int // Cash position category
	SharesCat int // Shares position category
}

const (
	// Portfolio position categories (3 categories: None, Medium, High)
	PosNone = iota
	PosMedium
	PosHigh
	NumPositionCategories = 3
)

const (
	// Market state space: 5040 MA ordering states (7! permutations)
	NumMarketStates = 5040
	// Total state space: MA states × cash categories × shares categories
	NumStates = NumMarketStates * NumPositionCategories * NumPositionCategories
)

// NewState creates a new State from component categories.
func NewState(maState, cashCat, sharesCat int) State {
	return State{
		Index:     Encode(maState, cashCat, sharesCat),
		MAState:   maState,
		CashCat:   cashCat,
		SharesCat: sharesCat,
	}
}

// Encode encodes (ma_state, cash_cat, shares_cat) into a single state index.
func Encode(maState, cashCat, sharesCat int) int {
	return maState*NumPositionCategories*NumPositionCategories + cashCat*NumPositionCategories + sharesCat
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
