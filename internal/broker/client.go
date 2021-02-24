package broker

import (
	"time"

	"pacabid/internal/stock"
)

type Client interface {
	CancelOrder(orderID string) error
	ExitAllPositions() error
	GetSymbolBars(symbol string, numMinutes int) ([]*stock.Bar, error)
	GetPosition(symbol string) (*stock.Position, error)
	ListOrders(orderType stock.OrderType, until time.Time) ([]*stock.Order, error)
	SubmitOrder(symbol string, quantity int, side stock.Side) error
	WaitForMarket() error
}
