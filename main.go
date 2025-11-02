package main

import (
	"exchange/orderbook"

	"github.com/valyala/fasthttp"
)

func main() {
	fasthttp.ListenAndServe(":3000", handlePlaceOrder)
}

type Market string

const (
	MaketETH Market = "ETH"
)

type Exchange struct {
	orderbooks map[Market]*orderbook.OrderBook
}

func NewExchange() *Exchange {
	return &Exchange{
		orderbooks: make(map[Market]*orderbook.OrderBook),
	}
}

func handlePlaceOrder(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(200)
	ctx.SetBodyString(`"hello world"`)
}
