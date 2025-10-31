package main

import (
	"fmt"
	"sort"
	"time"
)

type Match struct { // always match a bid w a Ask and vice versa
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

type Order struct {
	Size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
}

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Bid:       bid,
		Size:      size,
		Timestamp: time.Now().UnixNano(),
	}
}

func (o *Order) String() string {
	// String is a special function that is now being over ridden -- this is called inside fmt.Println
	return fmt.Sprintf("[size %.2f]", o.Size)

}

type Orders []*Order

func (o Orders) Len() int {
	return len(o)
}

func (o Orders) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o Orders) Less(i, j int) bool {
	return o[i].Timestamp < o[j].Timestamp
}

// a limit is basically a group of orders at a certain point
type Limit struct {
	Price       float64
	Orders      Orders
	TotalVolume float64
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

type Limits []*Limit
type ByBestAsk struct{ Limits }

func (b *ByBestAsk) Len() int {
	return len(b.Limits)
}

func (b *ByBestAsk) Swap(i, j int) {
	b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i]
}

func (b *ByBestAsk) Less(i, j int) bool {
	return b.Limits[i].Price < b.Limits[j].Price
}

type ByBestBid struct{ Limits }

func (b *ByBestBid) Len() int {
	return len(b.Limits)
}

func (b *ByBestBid) Swap(i, j int) {
	b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i]
}

func (b *ByBestBid) Less(i, j int) bool {
	return b.Limits[i].Price < b.Limits[j].Price // for bid high price is better hence opposite
}

func (l *Limit) AddOrder(o *Order) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}

func (l *Limit) DeleteOrder(o *Order) {
	// deleting the order also decreases the volume
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == o {
			l.Orders[i] = l.Orders[len(l.Orders)-1] // assigne the last elm to current
			l.Orders = l.Orders[:len(l.Orders)-1]   // truncate the last item - sorting is effectecd
			break
		}
	}

	o.Limit = nil           // the limit pointer should not remain dangling - satisfy the GC
	l.TotalVolume -= o.Size // decrease the volume
	sort.Sort(l.Orders)
}

type OrderBook struct {
	Asks []*Limit
	Bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Asks:      []*Limit{},
		Bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}

func (ob *OrderBook) PlaceOrder(price float64, o *Order) []Match {
	// when u place an order on the exchange 2 things happen
	// 1. it will match with minimun 1 other order -- bid or ask will be completed

	// 2. it does not match and then it sit in the books
	if o.Size > 0 {
		// if size is greater then 0 then we need to put the resting size in the books
		ob.add(price, o)
	}

	return []Match{}
}

func (ob *OrderBook) add(price float64, o *Order) {
	var limit *Limit

	if o.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}
	if limit == nil {
		// we initalise the limit
		limit = NewLimit(price)
		limit.AddOrder(o)
		if o.Bid {
			ob.Bids = append(ob.Bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.Asks = append(ob.Asks, limit)
			ob.AskLimits[price] = limit
		}
	}
}
