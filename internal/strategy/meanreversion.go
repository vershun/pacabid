package strategy

import (
	"fmt"
	"log"
	"os"
	"pacabid/internal/broker"
	"pacabid/internal/stock"
	"time"
)

type MeanReversion struct {
	client       broker.Client
	budget       float64
	numBars      int
	targetSymbol string

	lastOrder string
	closeLog  *log.Logger
}

func NewMeanReversion(symbol string, numBars int) *MeanReversion {
	ln := fmt.Sprintf("history/%s.log", time.Now().Format("2006-01-02"))
	f, err := os.OpenFile(ln, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	log.SetFlags(0)
	l := log.New(f, "", log.Ltime)
	return &MeanReversion{
		closeLog:     l,
		targetSymbol: symbol,
		numBars:      numBars,
	}
}

func (mr *MeanReversion) Prepare(budget float64, client broker.Client) {
	mr.budget = budget
	mr.client = client
}

func (mr *MeanReversion) Run(quit chan struct{}) error {
	if err := mr.client.ExitAllPositions(); err != nil {
		return fmt.Errorf("failed to exit all positions: %s", err)
	}
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	if err := mr.update(); err != nil {
		return err
	}
	for {
		select {
		case <-quit:
			return nil
		case <-t.C:
			if err := mr.update(); err != nil {
				fmt.Fprintln(os.Stderr, "ERROR: PANIC: ", err)
				//return err
			}
		}
	}
}

func (mr *MeanReversion) update() error {
	d, err := mr.client.TimeUntilMarketOpen()
	if err != nil {
		return err
	}
	if d > 0 {
		log.Printf("%d minutes until next market open.\n", int(d.Minutes()))
		return nil
	}

	bars, err := mr.client.GetSymbolBars(mr.targetSymbol, mr.numBars)
	if err != nil {
		return err
	}
	if len(bars) < mr.numBars {
		log.Printf("Waiting for %d more minutes to accumulate data.\n", mr.numBars-len(bars))
		return nil
	}

	positionCount := 0
	positionVal := 0.0
	position, err := mr.client.GetPosition(mr.targetSymbol)
	if err != nil {
		if err != broker.ErrPositionDoesNotExist {
			return fmt.Errorf("failed to get position for %s: %s", mr.targetSymbol, err)
		}
	}
	if position != nil {
		positionCount = position.Quantity
		positionVal = position.MarketValue
	}
	meanPrice := 0.0
	for _, b := range bars {
		meanPrice += float64(b.Close)
	}
	meanPrice /= float64(mr.numBars)
	curPrice := bars[len(bars)-1].Close
	if curPrice <= .001 {
		log.Println("ERROR IN CURRENT PRICE!! SET TO 0!")
		return nil
	}
	mr.closeLog.Printf("%f", curPrice)
	log.Printf("Mean price: %f, current price: %f.", meanPrice, curPrice)

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
		portfolioShare := (meanPrice - curPrice) / curPrice * 200
		log.Printf("Portfolio %f to stock.\n", portfolioShare)
		targetPositionVal := acct.PortfolioValue * portfolioShare
		amountToAdd := targetPositionVal - positionVal

		log.Printf("Adding %f to stock.\n", amountToAdd)

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
func (mr *MeanReversion) submitLimitOrder(symbol string, qty int, price float64, side stock.Side) error {
	if qty <= 0 {
		log.Printf("Order of | %d %s %s | not sent.", qty, symbol, side)
		return nil
	}
	if _, err := mr.client.SubmitLimitOrder(symbol, qty, price, side); err != nil {
		fmt.Printf("Order of | %d %s %s | did not go through.\n", qty, symbol, side)
		return err
	}
	fmt.Printf("Limit order of | %d %s %s | sent.\n", qty, symbol, side)
	return nil
}

// Submit a market order if quantity is above 0.
func (mr *MeanReversion) submitMarketOrder(symbol string, qty int, side stock.Side) error {
	if qty <= 0 {
		log.Printf("Order of | %d %s %s | not sent.", qty, symbol, side)
		return nil
	}
	if _, err := mr.client.SubmitMarketOrder(symbol, qty, side); err != nil {
		fmt.Printf("Order of | %d %s %s | did not go through.\n", qty, symbol, side)
		return err
	}
	fmt.Printf("Market order of | %d %s %s | sent.\n", qty, symbol, side)
	return nil
}

/*
func lwma(bars []*stock.Bar, alpha float64) float64 {

	return 0
}
*/
