package agent

// Action represents a portfolio rebalancing action.
type Action int

const (
	// ActionNothing - no rebalancing
	ActionNothing Action = iota
	// ActionBuySmall - buy small amount (10% of cash)
	ActionBuySmall
	// ActionBuyLarge - buy large amount (50% of cash)
	ActionBuyLarge
	// ActionSellSmall - sell small amount (10% of shares)
	ActionSellSmall
	// ActionSellLarge - sell large amount (50% of shares)
	ActionSellLarge
)

const NumActions = 5

// Action constants for portfolio fractions
const (
	BuySmall  = 0.1
	BuyLarge  = 0.5
	SellSmall = 0.1
	SellLarge = 0.5
)

// String returns a human-readable name for the action.
func (a Action) String() string {
	switch a {
	case ActionNothing:
		return "nothing"
	case ActionBuySmall:
		return "buy-small"
	case ActionBuyLarge:
		return "buy-large"
	case ActionSellSmall:
		return "sell-small"
	case ActionSellLarge:
		return "sell-large"
	default:
		return "unknown"
	}
}
