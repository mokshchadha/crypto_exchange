package orderbook

import (
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%v != %v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrder := NewOrder(true, 5) // bid is true then buy order
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)

	l.AddOrder(buyOrder)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.DeleteOrder(buyOrderB)

	fmt.Println(l)

}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderBook()
	sellOrder := NewOrder(false, 10)
	sellOrder2 := NewOrder(false, 5)
	ob.PlaceLimitOrder(10_000, sellOrder)
	ob.PlaceLimitOrder(9_000, sellOrder2)
	assert(t, len(ob.asks), 2)
	assert(t, len(ob.Orders), 2)
	assert(t, ob.Orders[sellOrder.ID], sellOrder)
	assert(t, ob.Orders[sellOrder2.ID], sellOrder2)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderBook()
	sellOrder := NewOrder(false, 20)
	ob.PlaceLimitOrder(10_000, sellOrder)
	buyOrder := NewOrder(true, 15)
	matches := ob.PlaceMarketOrder(buyOrder)
	assert(t, len(matches), 1)
	assert(t, len(ob.asks), 1)
	assert(t, ob.AskTotalVolume(), 5.0)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].SizeFilled, 15.0)
	assert(t, buyOrder.IsFilled(), true)
}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	ob := NewOrderBook()
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)
	buyOrderD := NewOrder(true, 1)

	// limit is a bucket of prices sitting at the same price of different sizes of same stock
	ob.PlaceLimitOrder(10_000, buyOrderA)
	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(5_000, buyOrderC)
	ob.PlaceLimitOrder(5_000, buyOrderD)

	assert(t, ob.BidTotalVolume(), 24.0)

	sellOrder := NewOrder(false, 20)
	matches := ob.PlaceMarketOrder(sellOrder)
	assert(t, ob.BidTotalVolume(), 4.0)
	assert(t, len(matches), 3) // limit is bucket of order at same price
	assert(t, len(ob.bids), 1)

}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderBook()
	buyOrder := NewOrder(true, 5)
	ob.PlaceLimitOrder(10_000, buyOrder)
	assert(t, ob.BidTotalVolume(), 5.0)
	assert(t, len(ob.Orders), 1)
	assert(t, ob.Orders[buyOrder.ID], buyOrder)

	ob.CancelOrder(buyOrder)
	assert(t, ob.BidTotalVolume(), 0.0)
	assert(t, len(ob.Orders), 0)

}
