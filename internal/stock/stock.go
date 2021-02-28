package stock

// Bar contains the information necessary to draw a bar in a bar chart.
type Bar struct {
	Close  float64
	High   float64
	Low    float64
	Open   float64
	Time   int64
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
	Quantity int
	Side     Side
	Symbol   string
}

type Position struct {
	CurrentPrice float64
	Exchange     string
	ID           string
	MarketValue  float64
	Quantity     int
	Side         Side
	Symbol       string
}
