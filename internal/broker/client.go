package broker

import (
	"time"

	"pacabid/internal/stock"
)

type Account struct {
	BuyingPower float64
	Cash        float64
}

type Client interface {
	CancelOrder(orderID string) error
	ExitAllPositions() error
	GetAccount() (*Account, error)
	GetPosition(symbol string) (*stock.Position, error)
	GetAllPositions() ([]*stock.Position, error)
	GetSymbolBars(symbol string, numMinutes int) ([]*stock.Bar, error)
	ListOrders(orderType stock.OrderType, until time.Time) ([]*stock.Order, error)
	SubmitMarketOrder(symbol string, quantity int, side stock.Side) error
	SubmitLimitOrder(symbol string, quantity int, price float64, side stock.Side) error
	WaitForMarket() error
}
