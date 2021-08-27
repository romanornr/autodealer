// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package webserver

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"net/http"
	"strings"

	"context"
)

type TickerResponse struct {
	ticker.Price
}

func (t TickerResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func TickerCtx(next http.Handler) http.Handler {
	base := new(Asset)
	quote := new(Asset)

	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		exchangeNameReq := chi.URLParam(request, "exchange")
		exchangeEngine, _ := engine.Bot.GetExchangeByName(exchangeNameReq)

		base.Code = currency.NewCode(strings.ToUpper(chi.URLParam(request, "base")))
		quote.Code = currency.NewCode(strings.ToUpper(chi.URLParam(request, "quote")))

		pair := currency.NewPair(base.Code, quote.Code)

		tickerInfo, err := exchangeEngine.FetchTicker(pair, asset.Spot)
		if err != nil {
			fmt.Printf("error %s\n", err)
		}

		ctx := context.WithValue(request.Context(), "tickerInfo", tickerInfo)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

func getTicker(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tickerInfo, ok := ctx.Value("tickerInfo").(*ticker.Price)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		render.Render(w, r, ErrNotFound)
		return
	}
	render.Render(w, r, TickerResponse{*tickerInfo})
	return
}
type ToUSDResponse struct {
	Price float64 `json:"price"`
}

func (t ToUSDResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func PriceToUSDCtx(next http.Handler) http.Handler {
	base := new(Asset)

	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		exchangeNameReq := chi.URLParam(request, "exchange")
		exchangeEngine, _ := engine.Bot.GetExchangeByName(exchangeNameReq)
		base.Code = currency.NewCode(strings.ToUpper(chi.URLParam(request, "base")))

		price, err := fetchTickerPriceUSD(exchangeEngine, base.Code)
		if err != nil {
			logrus.Errorf("failed to fetch price: %s\n", err)
		}
		logrus.Infof("price %s %f\n", base.Code, price)

		ctx := context.WithValue(request.Context(), "price", ToUSDResponse{Price: price})
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

func getUSDPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	price, ok := ctx.Value("price").(*ToUSDResponse)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		render.Render(w, r, ErrNotFound)
		return
	}
	render.Render(w, r, price)
}

func fetchPairQuotes(e exchange.IBotExchange, code currency.Code) ([]currency.Code, error) {
	pairs, err := e.FetchTradablePairs(asset.Spot)
	if err != nil {
		logrus.Errorf("exchange fetch tradble pairs: %s", err)
	}

	fmt.Println(code.String())
	// quote slice ["USD", "USDT", "BTC", "ETH"]
	var quotes []currency.Code
	for _, p := range pairs {
		availablePair := strings.Split(p, "/")
		fmt.Println(availablePair[0])
		if availablePair[0] == code.String() {
			quotes = append(quotes, currency.NewCode(availablePair[1]))
		}
	}
	logrus.Println(quotes)
	return quotes, err
}

// Fetch price ticker USD prices
// for example currencyPair VIA/BTC and there is no USD pair available
// check which quote pairs are available and if not check if ETH or BTC quote is available
func fetchTickerPriceUSD(e exchange.IBotExchange, code currency.Code) (float64, error) {
	var priceTicker *ticker.Price
	var err error
	var priceUSD float64

	var USD bool
	var USDT bool
	var BTC bool
	var ETH bool

	quotes, err := fetchPairQuotes(e, code)
	for _, q := range quotes {
		USD = q.Match(currency.USD)
		USDT = q.Match(currency.USDT)
		BTC = q.Match(currency.BTC)
		ETH = q.Match(currency.ETH)
	}

	if USD {
		pair := currency.NewPair(code, currency.USD)
		priceTicker, err = e.FetchTicker(pair, asset.Spot)
		priceUSD = priceTicker.Last
	}

	if USDT {
		pair := currency.NewPair(code, currency.USDT)
		priceTicker, err = e.FetchTicker(pair, asset.Spot)
		priceUSD = priceTicker.Last
	}

	if BTC {
		pair := currency.NewPair(code, currency.BTC)
		priceTicker, err = e.FetchTicker(pair, asset.Spot)

		priceTickerBTC, err := e.FetchTicker(currency.NewPair(currency.BTC, currency.USDT), asset.Spot)
		if err != nil {
			logrus.Errorf("failed to fetch ticker: %v", err)
		}
		priceUSD = priceTickerBTC.Last * priceTicker.Last
	}

	if ETH {
		pair := currency.NewPair(code, currency.BTC)
		priceTicker, err = e.FetchTicker(pair, asset.Spot)

		priceTickerETH, err := e.FetchTicker(currency.NewPair(currency.BTC, currency.USDT), asset.Spot)
		if err != nil {
			logrus.Errorf("failed to fetch ticker: %v", err)
		}
		priceUSD = priceTickerETH.Last * priceTicker.Last
	}

	if err != nil {
		logrus.Errorf("error %s\n", err)
	}

	return priceUSD, err

}
