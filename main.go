package main

import (
	"encoding/json"
	"exchange/orderbook"
	"fmt"
	"strconv"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func main() {
	ex := NewExchange()
	r := router.New()
	r.POST("/order", ex.handlePlaceOrder)
	r.POST("/cancel/{orderId}", ex.handleCancelOrder)
	r.GET("/books/{marketId}", ex.handleGetBook)
	fasthttp.ListenAndServe(":3000", r.Handler)
}

const invaliErrorStr = `{"error": "Invalid Error"}`

type OrderType string

const (
	MarketOrder OrderType = "MARKET"
	LimiteOrder OrderType = "LIMIT"
)

type Market string

const (
	MarketETH Market = "ETH"
)

type Exchange struct {
	orderbooks map[Market]*orderbook.OrderBook
}

func NewExchange() *Exchange {
	orderbooks := make(map[Market]*orderbook.OrderBook)
	orderbooks[MarketETH] = orderbook.NewOrderBook()

	return &Exchange{
		orderbooks: orderbooks,
	}
}

type PlaceOrderRequest struct {
	Type   OrderType // limit or market
	Bid    bool
	Size   float64
	Price  float64
	Market Market
}

type MatchedOrder struct {
	Price float64
	Size  float64
	ID    int64
}

func (ex *Exchange) handlePlaceOrder(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	body := ctx.PostBody()
	var placeOrderReq PlaceOrderRequest
	err := json.Unmarshal(body, &placeOrderReq)

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(invaliErrorStr)
		return
	}
	market := Market(placeOrderReq.Market)

	ob := ex.orderbooks[market]
	order := orderbook.NewOrder(placeOrderReq.Bid, placeOrderReq.Size)
	if placeOrderReq.Type == LimiteOrder {
		ob.PlaceLimitOrder(placeOrderReq.Price, order) // sell or buy at a particular price (lower or higher) is limit order
	} else {
		matches := ob.PlaceMarketOrder(order) // execute immediately on best available price (hence price is not needed cause it is decided by the market)
		matchedOrders := make([]*MatchedOrder, len(matches))

		isBid := false
		if order.Bid {
			isBid = true
		}

		for i := 0; i < len(matches); i++ {

			matchId := matches[i].Bid.ID
			if isBid {
				matchId = matches[i].Ask.ID
			}
			matchedOrders[i] = &MatchedOrder{
				Size:  matches[i].SizeFilled,
				Price: matches[i].Price,
				ID:    matchId,
			}
		}
		jsonData, err := json.Marshal(matchedOrders)
		if err != nil {
			fmt.Println("Could not marshal the matchedOrders")
		}
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBody(jsonData)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString(`{"message": "Order placed successfully"}`)
}

type Order struct {
	ID        int64
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}

type OrderBookData struct {
	TotalBidVolume  float64
	TotalAskVolumne float64
	Asks            []*Order
	Bids            []*Order
}

func (ex *Exchange) handleGetBook(ctx *fasthttp.RequestCtx) {
	marketId := Market(ctx.UserValue("marketId").(string))

	ob, ok := ex.orderbooks[marketId]

	if !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(invaliErrorStr)
		return
	}

	orderbookData := OrderBookData{
		Asks:            []*Order{},
		Bids:            []*Order{},
		TotalBidVolume:  ob.BidTotalVolume(),
		TotalAskVolumne: ob.AskTotalVolume(),
	}

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := Order{
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}

			orderbookData.Asks = append(orderbookData.Asks, &o)
		}

	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := Order{
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}

			orderbookData.Bids = append(orderbookData.Bids, &o)
		}

	}

	// instead send the orderbook data
	jsonData, err := json.Marshal(orderbookData)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(invaliErrorStr)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(jsonData)
}

type CancelOrderRequest struct {
}

func (ex *Exchange) handleCancelOrder(ctx *fasthttp.RequestCtx) {
	orderIdStr := ctx.UserValue("orderId").(string)
	orderId, err := strconv.ParseInt(orderIdStr, 10, 64)

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error": "Invalid order ID"}`)
		return
	}

	ob, ok := ex.orderbooks[MarketETH]

	if !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error": "Order book not found"}`)
		return
	}

	order := ob.Orders[orderId]
	ob.CancelOrder(order)

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString(`{"msg": "Order deleted successfully"}`)
}
