package broker

import (
	"fmt"
	"log"
	"pacabid/internal/stock"
	"time"

	alp "github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/shopspring/decimal"
)

type alpaca struct {
	client *alp.Client
}

/*
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
*/

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

func (a *alpaca) WaitToStart() error {
	log.Println("Waiting for market to open.")
	for {
		clock, err := a.client.GetClock()
		if err != nil {
			return err
		}
		if clock.IsOpen {
			break
		}
		timeToOpen := time.Duration(clock.NextOpen.Sub(clock.Timestamp).Minutes())
		log.Printf("%d minutes until next market open.\n", timeToOpen)
		time.Sleep((timeToOpen/2 + 1) * time.Minute)
	}
	log.Println("Market is open!")

	return nil
}

func (a *alpaca) toBar(alpBar *alp.Bar) *stock.Bar {
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

func (a *alpaca) toPosition(alpPos *alp.Position) *stock.Position {
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
		bars = append(bars, a.toBar(&b))
	}
	return bars, nil
}

func (a *alpaca) GetPositions() ([]*stock.Position, error) {
	alpPos, err := alp.client.GetPosition(alpacaClient.stock)
	panic("implement")
	return nil, nil
}

func (a *alpaca) GetPosition(symbol string) (*stock.Position, error) {
	alpPos, err := alp.client.GetPosition(alpacaClient.stock)
	panic("implement")
	return nil, nil
}

func (a *alpaca) ListOrders(orderType stock.OrderType, until time.Time) ([]*stock.Order, error) {
	panic("implement")
}

func (a *alpaca) GetAccount() (*Account, error) {
	alpAcct, err := a.client.GetAccount()
	if err != nil {
		return nil, err
	}
	var acct Account
	acct.BuyingPower, _ = alpAcct.BuyingPower.Float64()
	acct.Cash, _ = alpAcct.Cash.Float64()
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
