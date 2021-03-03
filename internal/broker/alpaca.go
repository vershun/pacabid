package broker

import (
	"fmt"
	"pacabid/internal/stock"
	"time"

	alp "github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/shopspring/decimal"
)

type alpaca struct {
	client *alp.Client
}

func NewAlpaca() *alpaca {
	alp.SetBaseUrl("https://paper-api.alpaca.markets")
	return &alpaca{
		client: alp.NewClient(common.Credentials()),
	}
}

func (a *alpaca) ExitAllPositions() error {
	open, until, limit := "open", time.Now(), 1000
	orders, err := a.client.ListOrders(&open, &until, &limit, nil)
	if err != nil {
		return fmt.Errorf("could not list orders: %s", err)
	}
	for _, order := range orders {
		if err := a.client.CancelOrder(order.ID); err != nil {
			return fmt.Errorf("could not cancel order %s: %s", order.ID, err)
		}
	}
	return nil
}

func (a *alpaca) TimeUntilMarketOpen() (time.Duration, error) {
	clock, err := a.client.GetClock()
	if err != nil {
		return 0, err
	}
	if clock.IsOpen {
		return 0, nil
	}
	return clock.NextOpen.Sub(clock.Timestamp), nil
}

func toBar(alpBar *alp.Bar) *stock.Bar {
	if alpBar == nil {
		return nil
	}
	return &stock.Bar{
		Time:   alpBar.Time,
		Open:   float64(alpBar.Open),
		High:   float64(alpBar.High),
		Low:    float64(alpBar.Low),
		Close:  float64(alpBar.Close),
		Volume: alpBar.Volume,
	}
}

func toPosition(alpPos *alp.Position) *stock.Position {
	if alpPos == nil {
		return nil
	}
	p := &stock.Position{
		Exchange: alpPos.Exchange,
		ID:       alpPos.AssetID,
		Quantity: int(alpPos.Qty.IntPart()),
		Side:     stock.Side(alpPos.Side),
		Symbol:   alpPos.Symbol,
	}
	p.CurrentPrice, _ = alpPos.CurrentPrice.Float64()
	p.MarketValue, _ = alpPos.MarketValue.Float64()
	return p
}

func (a *alpaca) GetSymbolBars(symbol string, numMinutes int) ([]*stock.Bar, error) {
	alpBars, err := a.client.GetSymbolBars(
		symbol,
		alp.ListBarParams{Timeframe: "minute", Limit: &numMinutes},
	)
	if err != nil {
		return nil, err
	}
	var bars []*stock.Bar
	for _, b := range alpBars {
		bars = append(bars, toBar(&b))
	}
	return bars, nil
}

func (a *alpaca) GetPositions() ([]*stock.Position, error) {
	ps, err := a.client.ListPositions()
	if err != nil {
		return nil, err
	}
	var positions []*stock.Position
	for _, p := range ps {
		positions = append(positions, toPosition(&p))
	}
	return positions, nil
}

func (a *alpaca) GetPosition(symbol string) (*stock.Position, error) {
	p, err := a.client.GetPosition(symbol)
	if err != nil {
		if err.Error() == ErrPositionDoesNotExist.Error() {
			return nil, ErrPositionDoesNotExist
		}
		return nil, err
	}
	return toPosition(p), nil
}

func (a *alpaca) ListOrders(orderType stock.OrderType, until time.Time) ([]*stock.Order, error) {
	status, limit := "open", 1000
	os, _ := a.client.ListOrders(&status, &until, &limit, nil)
	var orders []*stock.Order
	for _, o := range os {
		orders = append(orders, &stock.Order{
			Quantity: int(o.Qty.IntPart()),
			Side:     stock.Side(o.Side),
			Symbol:   o.Symbol,
		})
	}
	return orders, nil
}

func (a *alpaca) GetAccount() (*Account, error) {
	alpAcct, err := a.client.GetAccount()
	if err != nil {
		return nil, err
	}
	var acct Account
	acct.BuyingPower, _ = alpAcct.BuyingPower.Float64()
	acct.Cash, _ = alpAcct.Cash.Float64()
	acct.PortfolioValue, _ = alpAcct.PortfolioValue.Float64()
	acct.ID = alpAcct.ID
	return &acct, nil
}

func (a *alpaca) SubmitMarketOrder(symbol string, quantity int, side stock.Side) (string, error) {
	acct, err := a.GetAccount()
	if err != nil {
		return "", nil
	}
	ord, err := a.client.PlaceOrder(
		alp.PlaceOrderRequest{
			AccountID:   acct.ID,
			AssetKey:    &symbol,
			Qty:         decimal.NewFromFloat(float64(quantity)),
			Side:        alp.Side(side),
			Type:        "market",
			TimeInForce: "day",
		})
	if err != nil {
		return "", err
	}
	return ord.ID, nil

}

func (a *alpaca) SubmitLimitOrder(symbol string, quantity int, price float64, side stock.Side) (string, error) {
	acct, err := a.GetAccount()
	if err != nil {
		return "", nil
	}
	decPrice := decimal.NewFromFloat(price)
	ord, err := a.client.PlaceOrder(
		alp.PlaceOrderRequest{
			AccountID:   acct.ID,
			AssetKey:    &symbol,
			Qty:         decimal.NewFromFloat(float64(quantity)),
			Side:        alp.Side(side),
			Type:        "limit",
			LimitPrice:  &decPrice,
			TimeInForce: "day",
		})
	if err != nil {
		return "", err
	}
	return ord.ID, nil
}

func (a *alpaca) CancelOrder(orderID string) error {
	return a.client.CancelOrder(orderID)
}
