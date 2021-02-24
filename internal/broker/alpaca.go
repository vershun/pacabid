package broker

import (
	"fmt"
	"log"
	"time"

	alp "github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
)

type alpaca struct {
	client *alp.Client
}

/*
type Client interface {
	CancelOrder(orderID string) error
	ExitAllPositions() error
	GetSymbolBars(symbol string, numMinutes int) ([]*stock.Bar, error)
	GetPositions(symbol string) (*Position, error)
	ListOrders(orderType stock.OrderType, until time.Time) ([]*stock.Order, error)
	SubmitOrder(symbol string, quantity int, side stock.Side) error
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
		fmt.Printf("%d minutes until next market open.\n", timeToOpen)
		time.Sleep((timeToOpen/2 + 1) * time.Minute)
	}
	fmt.Println("Market is open!")

	return nil
}
