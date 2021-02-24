package strategy

import (
	"log"
	"pacabid/internal/broker"
	"pacabid/internal/stock"
	"time"
)

const numBars = 25

type MeanRevision struct {
	budget       float64
	targetSymbol string
	client       broker.Client
}

func (mr *MeanRevision) Prepare(budget float64, client broker.Client) {
	mr.budget = budget
	mr.client = client
}

func (mr *MeanRevision) Run(quit chan struct{}) error {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-quit:
			return nil
		case <-t.C:
			if err := mr.update(); err != nil {
				return err
			}
		}
	}
}

func (mr *MeanRevision) update() error {
	numStock, val := 0, 0.0
	position, err := mr.client.GetPosition(mr.targetSymbol)
	if err != nil {
		return err
	}
	if position != nil {
		numStock = position.Quantity
		val = position.MarketValue
	}

	bars, err := mr.client.GetSymbolBars(mr.targetSymbol, numBars)
	if err != nil {
		return err
	}
	meanPrice := 0.0
	for _, b := range bars {
		meanPrice += float64(b.Close)
	}
	meanPrice /= float64(numBars)
	curPrice := bars[len(bars)-1].Close

	if curPrice > meanPrice {
		// We're above the mean! Sell if we have any.
		if numStock > 0 {
			log.Printf("Above the mean. Selling %d %s stock for %f.", numStock, mr.targetSymbol, position.MarketValue)
			return mr.client.SubmitOrder(mr.targetSymbol, numStock, stock.Sell); err != nil {
		}
	}




	return nil
}

/*
func lwma(bars []*stock.Bar, alpha float64) float64 {

	return 0
}
*/
