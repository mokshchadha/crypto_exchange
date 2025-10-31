package main

import (
	"fmt"
	"testing"
)

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

func TestOrderBook(t *testing.T) {
	ob := NewOrderBook()
	buyOrder := NewOrder(true, 10)
	buyOrder2 := NewOrder(true, 200)

	ob.PlaceOrder(18_000, buyOrder)
	ob.PlaceOrder(14_000, buyOrder2)

	for i := 0; i < len(ob.Bids); i++ {
		fmt.Printf("%+v", ob.Bids[i])
	}
}
