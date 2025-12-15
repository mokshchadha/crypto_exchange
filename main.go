package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"exchange/orderbook"
	"fmt"
	"log"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type (
	OrderType         string
	Market            string
	PlaceOrderRequest struct {
		Type   OrderType // limit or market
		Bid    bool
		Size   float64
		Price  float64
		Market Market
		UserId int64
	}

	MatchedOrder struct {
		Price float64
		Size  float64
		ID    int64
	}

	Order struct {
		ID        int64
		Price     float64
		Size      float64
		Bid       bool
		Timestamp int64
	}
)

const (
	MarketOrder    OrderType = "MARKET"
	LimiteOrder    OrderType = "LIMIT"
	MarketETH      Market    = "ETH"
	exchangePvtKey           = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d" // this should be env variable
	invaliErrorStr           = `{"error": "Invalid Error"}`
)

func main() {
	ganacheServer := "http://127.0.0.1:8545"
	client, err := ethclient.Dial(ganacheServer)

	if err != nil {
		log.Fatal("Some error while connecting with ganache")
	}
	ex := NewExchange(client)
	r := router.New()
	r.POST("/order", ex.handlePlaceOrder)
	r.POST("/cancel/{orderId}", ex.handleCancelOrder)
	r.GET("/books/{marketId}", ex.handleGetBook)

	// ctx := context.Background()
	// firstAddress := "0x90F8bf6A479f320ead074411a4B0e7944Ea8c9C1" // update this as per the ganache server
	// address := common.HexToAddress(firstAddress)

	// privateKey, err := crypto.HexToECDSA("4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d") // remove the 0x from the ganche pvt key

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// publicKey := privateKey.Public()
	// publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)

	// if !ok {
	// 	log.Fatal("error casting public key to ECDSA")
	// }

	// fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	// nonce, err := client.PendingNonceAt(context.Background(), fromAddress)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// value := big.NewInt(1000000000000000000) // in wei (1 eth)
	// gasLimit := uint64(21000)                //

	// gasPrice, err := client.SuggestGasPrice(context.Background())

	// toAddress := common.HexToAddress("0x1dF62f291b2E969fB0849d99D9Ce41e2F137006e")

	// tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
	// fmt.Println(tx)

	// // chainID, err := client.NetworkID(context.Background())
	// chainID := big.NewInt(1337)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)

	// if err != nil {
	// 	log.Fatal(err)

	// }
	// err = client.SendTransaction(context.Background(), signedTx)

	// balance, err := client.BalanceAt(ctx, address, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(balance)

	fasthttp.ListenAndServe(":3000", r.Handler)
}

type User struct {
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(privateKey string) *User {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic(err)
	}
	userPtr := &User{
		PrivateKey: pk,
	}
	return userPtr
}

type Exchange struct {
	Client     *ethclient.Client
	orderbooks map[Market]*orderbook.OrderBook
	PrivateKey *ecdsa.PrivateKey
	orders     map[int64]int64
	users      map[int64]*User
}

func NewExchange(client *ethclient.Client) *Exchange {
	orderbooks := make(map[Market]*orderbook.OrderBook)
	orderbooks[MarketETH] = orderbook.NewOrderBook()
	privateKey, err := crypto.HexToECDSA(exchangePvtKey) // remove the 0x from the ganche pvt key

	if err != nil {
		log.Fatal(err)
	}

	return &Exchange{
		Client:     client,
		orderbooks: orderbooks,
		PrivateKey: privateKey,
		users:      make(map[int64]*User),
		orders:     make(map[int64]int64),
	}
}

func (ex *Exchange) handlePlaceMarketOrder(market Market, o *orderbook.Order) ([]orderbook.Match, []*MatchedOrder) {
	ob := ex.orderbooks[market]
	matches := ob.PlaceMarketOrder(o) // execute immediately on best available price (hence price is not needed cause it is decided by the market)
	matchedOrders := make([]*MatchedOrder, len(matches))

	isBid := false
	if o.Bid {
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

	return matches, matchedOrders
}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {
	// for _, match := range matches {
	// 	// update the wallet if the matches are found
	// }
	return nil
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, o *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, o)
	user := ex.users[o.UserId]
	exchangePubKey := ex.PrivateKey.Public()
	publicKeyECDSA, ok := exchangePubKey.(*ecdsa.PublicKey)

	if !ok {
		return fmt.Errorf("cannot cast public key to ECDSA")
	}

	toAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	value := big.NewInt(1000000000000000000)
	err := transferETH(ex.Client, toAddress, user.PrivateKey, value)
	return err
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

	order := orderbook.NewOrder(placeOrderReq.Bid, placeOrderReq.Size, placeOrderReq.UserId)
	if placeOrderReq.Type == LimiteOrder {
		ex.handlePlaceLimitOrder(market, placeOrderReq.Price, order) // sell or buy at a particular price (lower or higher) is limit order
	} else {
		matches, matchedOrders := ex.handlePlaceMarketOrder(market, order)
		jsonData, err := json.Marshal(matchedOrders)
		if err != nil {
			fmt.Println("Could not marshal the matchedOrders")
		}
		fmt.Println(matches)
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBody(jsonData)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString(`{"message": "Order placed successfully"}`)
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
