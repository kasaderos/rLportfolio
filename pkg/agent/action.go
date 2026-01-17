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

// IsBuy returns true if the action is a buy action.
func (a Action) IsBuy() bool {
	return a == ActionBuySmall || a == ActionBuyLarge
}

// IsSell returns true if the action is a sell action.
func (a Action) IsSell() bool {
	return a == ActionSellSmall || a == ActionSellLarge
}

// IsTrade returns true if the action is a buy or sell (not nothing).
func (a Action) IsTrade() bool {
	return a.IsBuy() || a.IsSell()
}

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
