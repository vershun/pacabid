package main

import (
	"fmt"
	"log"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
)

type pacabid struct {
	client     *alpaca.Client
	strategies *strategies
}

type strategies struct {
	strategies []tradingStrategy
}

type tradingStrategy interface {
	run() error
	setWeight(weight int)
	weight() int
}

type strategy struct {
	weight      int
	percOfTotal float64
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
	open, until, limit := "open", time.Now(), 100
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
	fmt.Println("Market is open! Starting pacabid.")

	return nil
}

func main() {
	pb := &pacabid{
		client: alpaca.NewClient(common.Credentials()),
	}

	if err := pb.exitAllPositions(); err != nil {
		log.Fatal("Failed to exit all positions:", err)
	}

	if err := pb.waitToStart(); err != nil {
		log.Fatal("Failed waiting for market to open:", err)
	}

	/*
		acct, err := pb.client.GetAccount()
		if err != nil {
			log.Fatal("Failed to get Alpaca account:", err)
		}
	*/
}
