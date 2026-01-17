package state

// State represents the market state for portfolio optimization.
// It encodes moving average ordering state, convergence/divergence, and portfolio position.
type State struct {
	// Encoded state index (0 to numStates-1)
	Index int

	// Raw components (for debugging/analysis)
	MAState      int // Moving average ordering state (0-5039)
	MADivergence int // MA convergence/divergence: 0=converging, 1=neutral, 2=diverging
	CashCat      int // Cash position category
	SharesCat    int // Shares position category
}

const (
	// Portfolio position categories (3 categories: None, Medium, High)
	PosNone = iota
	PosMedium
	PosHigh
	NumPositionCategories = 3
)

const (
	// MA divergence categories
	MAConverging              = iota // MAs are getting closer together
	MANeutral                        // MAs spread is stable
	MADiverging                      // MAs are getting farther apart
	NumMADivergenceCategories = 3
)

const (
	// Market state space: 5040 MA ordering states (7! permutations)
	NumMarketStates = 5040
	// Total state space: MA states × MA divergence × cash categories × shares categories
	NumStates = NumMarketStates * NumMADivergenceCategories * NumPositionCategories * NumPositionCategories
)

// NewState creates a new State from component categories.
func NewState(maState, maDivergence, cashCat, sharesCat int) State {
	return State{
		Index:        Encode(maState, maDivergence, cashCat, sharesCat),
		MAState:      maState,
		MADivergence: maDivergence,
		CashCat:      cashCat,
		SharesCat:    sharesCat,
	}
}

// Encode encodes (ma_state, ma_divergence, cash_cat, shares_cat) into a single state index.
func Encode(maState, maDivergence, cashCat, sharesCat int) int {
	maStateWithDivergence := maState*NumMADivergenceCategories + maDivergence
	return maStateWithDivergence*NumPositionCategories*NumPositionCategories + cashCat*NumPositionCategories + sharesCat
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
