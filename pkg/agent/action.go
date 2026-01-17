package agent

// Action represents a portfolio rebalancing action.
type Action int

const (
	// ActionNothing - no rebalancing
	ActionNothing Action = iota
	// ActionBuySmall - buy small amount (10% of cash)
	ActionBuySmall
	// ActionBuyMedium - buy medium amount (30% of cash)
	ActionBuyMedium
	// ActionBuyLarge - buy large amount (50% of cash)
	ActionBuyLarge
	// ActionSellSmall - sell small amount (10% of shares)
	ActionSellSmall
	// ActionSellMedium - sell medium amount (30% of shares)
	ActionSellMedium
	// ActionSellLarge - sell large amount (50% of shares)
	ActionSellLarge
)

const NumActions = 7

// Action constants for portfolio fractions
const (
	BuySmall   = 0.1
	BuyMedium  = 0.3
	BuyLarge   = 0.5
	SellSmall  = 0.1
	SellMedium = 0.3
	SellLarge  = 0.5
)

// String returns a human-readable name for the action.
func (a Action) String() string {
	switch a {
	case ActionNothing:
		return "nothing"
	case ActionBuySmall:
		return "buy-small"
	case ActionBuyMedium:
		return "buy-medium"
	case ActionBuyLarge:
		return "buy-large"
	case ActionSellSmall:
		return "sell-small"
	case ActionSellMedium:
		return "sell-medium"
	case ActionSellLarge:
		return "sell-large"
	default:
		return "unknown"
	}
}
