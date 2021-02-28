package broker

import (
	"time"

	"pacabid/internal/stock"
)

type Account struct {
	ID          string
	BuyingPower float64
	Cash        float64
}

type Client interface {
	CancelOrder(orderID string) error
	ExitAllPositions() error
	GetAccount() (*Account, error)
	GetPosition(symbol string) (*stock.Position, error)
	GetPositions() ([]*stock.Position, error)
	GetSymbolBars(symbol string, numMinutes int) ([]*stock.Bar, error)
	ListOrders(orderType stock.OrderType, until time.Time) ([]*stock.Order, error)
	TimeUntilMarketOpen() (time.Duration, error)

	// SubmitMarketOrder submits an order to buy or sell quantity of a symbol
	// at the current market price.
	SubmitMarketOrder(symbol string, quantity int, side stock.Side) (string, error)

	// SubmitLimitOrder submits an order to buy or sell quantity of a symbol
	// at the set price.
	SubmitLimitOrder(symbol string, quantity int, price float64, side stock.Side) (string, error)
}
