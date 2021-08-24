package maker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/romanornr/autodealer/algo"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

type Token struct {
	Address           string `json:"address"`
	Decimals          string `json:"decimals"`
	Name              string `json:"name"`
	Symbol            string `json:"symbol"`
	TotalSupply       string `json:"totalSupply"`
	TransfersCount    int    `json:"transfersCount"`
	LastUpdated       int    `json:"lastUpdated"`
	IssuancesCount    int    `json:"issuancesCount"`
	HoldersCount      int    `json:"holdersCount"`
	EthTransfersCount int    `json:"ethTransfersCount"`
	Price             struct {
		Rate            float64 `json:"rate"`
		Diff            float64 `json:"diff"`
		Diff7D          float64 `json:"diff7d"`
		Ts              int     `json:"ts"`
		MarketCapUsd    float64 `json:"marketCapUsd"`
		AvailableSupply int     `json:"availableSupply"`
		Volume24H       float64 `json:"volume24h"`
		VolDiff1        float64 `json:"volDiff1"`
		VolDiff7        float64 `json:"volDiff7"`
		Currency        string  `json:"currency"`
	} `json:"price"`
	CountOps int `json:"countOps"`
}

func GetEthereumToken() *Token {
	logrus.Info("Get external price")
	apiEndpoint, err := url.Parse("https://api.ethplorer.io/getTokenInfo/0xb753428af26e81097e7fd17f40c88aaa3e04902c?apiKey=freekey")
	if err != nil {
		logrus.Fatalf("could not parse url endpoint: %s\n", err)
	}

	client := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       time.Second * 15,
	}

	req, err := http.NewRequest(http.MethodGet, apiEndpoint.String(), nil)
	if err != nil {
		logrus.Fatalf("could not create a new request to %s\n", apiEndpoint.Scheme)
	}

	resp, err := client.Do(req)
	if err != nil {
		logrus.Fatalf("client do failed for API endpoint: %s\n", apiEndpoint.String())
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Fatalf("Could not read body response from API endpoint %s %s\n", apiEndpoint.String(), err)
	}

	tokens := new(Token)
	err = json.Unmarshal(body, &tokens)

	return tokens
}

func btseGrid(bidLimitPrice, stepsize float64) {
	currencyPair, err := currency.NewPairFromString("SFI-USDT")
	if err != nil {
		logrus.Warnf("failed to retrieve trading pair: %s\n", currencyPair)
	}

	token := GetEthereumToken()
	externalPrice := token.Price.Rate

	exchangeEngine := engine.Bot.GetExchangeByName("Btse")
	ticker, err := exchangeEngine.FetchTicker(currencyPair, asset.Spot)
	if err != nil {
		logrus.Warn(err)
	}
	logrus.Infof("ticker bid: %f\n", ticker.Bid)

	diff := 100 / externalPrice * ticker.Bid
	fmt.Printf("Price difference is %f%%\n", 100-diff)

	//if bidLimitPrice == 0 {
	//	bidLimitPrice = externalPrice / 100 * (100 - 3) // 3% below external price
	//}

	logrus.Info("start placing grids")

	go func() {
		orders := gridOrders(1, 6, 100, 488.00, stepsize, order.Buy, currencyPair)
		executeMultipleOrdersAtOnce(exchangeEngine, orders)
		logrus.Printf("%s orders added to the orderbook\n", orders[0].Side.Lower())
	}()

	ordersReverse := gridOrders(3, 15, 300, 510, stepsize+1, order.Sell, currencyPair)
	executeMultipleOrdersAtOnce(exchangeEngine, ordersReverse)
}

func gridOrders(dollarSizeMin, dollarSizeMax, amountOfOrders, quoteLimitPrice float64, stepsize float64, side order.Side, pair currency.Pair) []order.Submit {
	var orders = []order.Submit{}

	rand.Seed(time.Now().UnixNano())
	var x = algo.RandFloats(dollarSizeMin, dollarSizeMax, int(amountOfOrders))

	var gridPrice = quoteLimitPrice
	for i := 0.0; i < amountOfOrders; i++ {

		if side == order.Buy {
			gridPrice = gridPrice - stepsize
		}

		if side == order.Sell {
			gridPrice += stepsize
		}

		amount := x[int(i)] / gridPrice
		amount = roundTo(amount, 3)

		o := order.Submit{
			PostOnly:  true,
			Price:     roundTo(gridPrice, 1),
			Amount:    amount,
			Exchange:  "Btse",
			AssetType: asset.Spot,
			Type:      order.Limit,
			Side:      side,
			Pair:      pair,
		}

		if o.Side == order.Buy && o.Price >= 480 {
			fmt.Printf("cancel price too high %f\n", o.Price)
			continue
		}

		orders = append(orders, o)
	}
	return orders
}

func roundTo(n float64, decimals uint32) float64 {
	return math.Round(n*math.Pow(10, float64(decimals))) / math.Pow(10, float64(decimals))
}

func executeMultipleOrdersAtOnce(e exchange.IBotExchange, orders []order.Submit) {
	for _, o := range orders {
		_, err := e.SubmitOrder(&o)
		if err != nil {
			o.Price -= 1.0 // retry?
			_, err = e.SubmitOrder(&o)
			if err != nil {
				logrus.Warnf("Failed second try submitting order %s %f at %f: %s\n", o.Pair, o.Amount, o.Price, err)
			}
		}
		fmt.Printf("%s Amount: %f Price: %f\n", strings.ToLower(o.Side.Lower()), o.Amount, o.Price)
		time.Sleep(time.Millisecond * 50)
	}
}
