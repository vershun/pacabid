package stock

// Bar contains the information necessary to draw a bar in a bar chart.
type Bar struct {
	Time   int64
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int32
}

type Side string
type OrderType string

const (
	Buy    = Side("buy")
	Sell   = Side("sell")
	Limit  = OrderType("limit")
	Market = OrderType("market")
)

// Order describes an execution option on a market.
type Order struct {
	Quantity     int
	Side         Side
	StrategyName string
	Symbol       string
}

type Position struct {
	ID           string
	Symbol       string
	Exchange     string
	Quantity     int
	Side         Side
	MarketValue  float64
	CurrentPrice float64
}
