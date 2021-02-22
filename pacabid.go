package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
)

type side string
type orderType string

const (
	buy    = side("buy")
	sell   = side("sell")
	limit  = orderType("limit")
	market = orderType("market")
)

type pacabid struct {
	client       *alpaca.Client
	strategies   []tradingStrategy
	orderQueue   chan order
	marketStream chan bar
	quit         chan struct{}
}

type bar struct {
	time   int64
	open   float32
	high   float32
	low    float32
	close  float32
	volume int32
}

type order struct {
	strategyName string
	symbol       string
	quantity     int
	side         side
}

type client interface {
	symbolBars(symbol string, numMinutes int) ([]*bar, error)
	submitOrder(symbol string, quantity int, side side) error
	cancelOrder(orderID string) error
	listOrders(orderType orderType, until time.Time) ([]*order, error)
	waitForMarket() error
}

type tradingStrategy interface {
	symbolsToWatch() []string
	prepare(budget float64, marketStream <-chan bar, orders chan<- order)
	run() error
}

type strategy struct {
	budgetMicros int64
}

/*
func (ss *strategies) rebalance() {
	var total int
	for _, s := range ss.strategies {
		total += s.weight()
	}
	for _, s := range ss.strategies {
		//percOfTotal += s.weight()
	}

}
*/

func (p *pacabid) exitAllPositions() error {
	open, until, limit := "open", time.Now(), 1000
	orders, err := p.client.ListOrders(&open, &until, &limit, nil)
	if err != nil {
		return fmt.Errorf("could not list orders: %s", err)
	}
	for _, order := range orders {
		if err := p.client.CancelOrder(order.ID); err != nil {
			return fmt.Errorf("could not cancel order %s: %s", order.ID, err)
		}
	}
	return nil
}

func (p *pacabid) waitToStart() error {
	log.Println("Waiting for market to open.")
	for {
		clock, err := p.client.GetClock()
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

func newPacabid() *pacabid {
	alpaca.SetBaseUrl("https://paper-api.alpaca.markets")
	return &pacabid{
		client:       alpaca.NewClient(common.Credentials()),
		marketStream: make(chan bar),
		orderQueue:   make(chan order, 100),
		quit:         make(chan struct{}),
	}
}

// Make this timer tick for quit chan
func (pb *pacabid) loopBarFeed() error {
	for {
		// Figure out when the market will close so we can prepare to sell beforehand.
		clock, err := pb.client.GetClock()
		if err != nil {
			return err
		}
		if clock.NextClose.Sub(clock.Timestamp) < 15*time.Minute {
			fmt.Println("Market closing soon.  Closing positions.")

			positions, _ := pb.client.ListPositions()
			for _, position := range positions {
				var orderSide string
				if position.Side == "long" {
					orderSide = "sell"
				} else {
					orderSide = "buy"
				}
				qty, _ := position.Qty.Float64()
				qty = math.Abs(qty)

				_ = orderSide
				//pb.client.submitMarketOrder(int(qty), position.Symbol, orderSide)
			}
			fmt.Println("Exiting.")
			return nil
		} else {
			limit := 1
			bs, err := pb.client.GetSymbolBars("AAPL", alpaca.ListBarParams{Timeframe: "minute", Limit: &limit})
			if err != nil {
				return err
			}
			b := bs[0]
			bar := bar{
				time:   b.Time,
				open:   b.Open,
				close:  b.Close,
				high:   b.High,
				low:    b.Low,
				volume: b.Volume,
			}
			// Send to listening strategies
			pb.marketStream <- bar
		}
		time.Sleep(1 * time.Minute)
	}
	return nil
}

func main() {
	pb := newPacabid()

	if err := pb.exitAllPositions(); err != nil {
		log.Fatal("Failed to exit all positions: ", err)
	}

	if err := pb.waitToStart(); err != nil {
		log.Fatal("Failed waiting for market to open: ", err)
	}

	/*
		acct, err := pb.client.GetAccount()
		if err != nil {
			log.Fatal("Failed to get Alpaca account:", err)
		}
	*/
}
