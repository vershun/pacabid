package strategy

import (
	"fmt"
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

	lastOrder string
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
	positionCount := 0
	positionVal := 0.0
	position, err := mr.client.GetPosition(mr.targetSymbol)
	if err != nil {
		return err
	}
	if position != nil {
		positionCount = position.Quantity
		positionVal = position.MarketValue
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
		if positionCount > 0 {
			log.Printf("Above the mean. Selling %d %s stock for %f.\n", positionCount, mr.targetSymbol, curPrice)
			return mr.submitLimitOrder(mr.targetSymbol, positionCount, curPrice, stock.Sell)
		}
	}
	if curPrice < meanPrice {
		acct, err := mr.client.GetAccount()
		if err != nil {
			return err
		}
		positions, err := mr.client.GetAllPositions()
		if err != nil {
			return err
		}
		var portfolioVal float64
		for _, p := range positions {
			portfolioVal += p.MarketValue
		}
		portfolioShare := (meanPrice - curPrice) / curPrice * 200
		targetPositionVal := portfolioVal * portfolioShare
		amountToAdd := targetPositionVal - positionVal

		if amountToAdd > 0 {
			if amountToAdd > acct.BuyingPower {
				amountToAdd = acct.BuyingPower
			}
			numToBuy := int(amountToAdd / curPrice)
			mr.submitLimitOrder(mr.targetSymbol, numToBuy, curPrice, stock.Buy)
			return nil
		}

		if amountToAdd < 0 {
			amountToAdd *= -1
			numToSell := int(amountToAdd / curPrice)
			if numToSell > positionCount {
				numToSell = positionCount
			}
			mr.submitLimitOrder(mr.targetSymbol, numToSell, curPrice, stock.Sell)
		}
	}
	return nil
}

// Submit a limit order if quantity is above 0.
func (mr *MeanRevision) submitLimitOrder(symbol string, qty int, price float64, side stock.Side) error {
	if qty <= 0 {
		log.Printf("Order of | %d %s %s | not sent.", qty, symbol, side)
		return nil
	}
	if err := mr.client.SubmitLimitOrder(symbol, qty, price, side); err != nil {
		fmt.Printf("Order of | %d %s %s | did not go through.\n", qty, symbol, side)
		return err
	}
	fmt.Printf("Limit order of | %d %s %s | sent.\n", qty, symbol, side)
	return nil
	/*
		order, err := alp.client.PlaceOrder(alpaca.PlaceOrderRequest{
			AccountID:   account.ID,
			AssetKey:    &symbol,
			Qty:         decimal.NewFromFloat(float64(qty)),
			Side:        adjSide,
			Type:        "limit",
			LimitPrice:  &limPrice,
			TimeInForce: "day",
		})
	*/
}

// Submit a market order if quantity is above 0.
func (mr *MeanRevision) submitMarketOrder(symbol string, qty int, side stock.Side) error {
	if qty <= 0 {
		log.Printf("Order of | %d %s %s | not sent.", qty, symbol, side)
		return nil
	}
	if err := mr.client.SubmitMarketOrder(symbol, qty, side); err != nil {
		fmt.Printf("Order of | %d %s %s | did not go through.\n", qty, symbol, side)
		return err
	}
	fmt.Printf("Market order of | %d %s %s | sent.\n", qty, symbol, side)
	return nil
	/*
		lastOrder, err := alp.client.PlaceOrder(alpaca.PlaceOrderRequest{
			AccountID:   account.ID,
			AssetKey:    &symbol,
			Qty:         decimal.NewFromFloat(float64(qty)),
			Side:        adjSide,
			Type:        "market",
			TimeInForce: "day",
		})
		if err == nil {
			fmt.Printf("Market order of | %d %s %s | completed.\n", qty, symbol, side)
		} else {
			fmt.Printf("Order of | %d %s %s | did not go through.\n", qty, symbol, side)
		}
		return err
		fmt.Printf("Quantity is <= 0, order of | %d %s %s | not sent.\n", qty, symbol, side)
		return nil
	*/
}

/*
func lwma(bars []*stock.Bar, alpha float64) float64 {

	return 0
}
*/
