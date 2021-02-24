package strategy

import "pacabid/internal/broker"

type TradingStrategy interface {
	Prepare(budget float64, client broker.Client)
	Run(quit chan struct{}) error
}
