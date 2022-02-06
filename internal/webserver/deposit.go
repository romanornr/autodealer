// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package webserver

import (
	"context"
	"errors"
	"github.com/romanornr/autodealer/internal/algo"
	"net/http"
	"strings"
	"time"

	"github.com/romanornr/autodealer/internal/dealer"
	"github.com/romanornr/autodealer/internal/singleton"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/deposit"
)

// depositResponse is the response payload for deposit requests
type depositResponse struct {
	Asset     *currency.Item   `json:"asset"`
	Code      currency.Code    `json:"code"`
	Chains    []string         `json:"chains"`
	Address   *deposit.Address `json:"address"`
	Time      time.Time        `json:"time"`
	Balance   float64          `json:"balance"`
	Price     float64          `json:"price"`
	Value     float64          `json:"value"`
	Err       error            `json:"error"`
	AccountID string           `json:"account"`
}

// DepositHandler handles deposit requests
func DepositHandler(w http.ResponseWriter, _ *http.Request) {
	err := tpl.ExecuteTemplate(w, "deposit.html", nil)
	if err != nil {
		logrus.Errorf("error template: %s\n", err)
	}
}

// getDepositAddress is a function that returns the deposit address for a given exchange and asset.
func getDepositAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	depositResponse, ok := ctx.Value("response").(*depositResponse)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.Status(r, http.StatusUnprocessableEntity)
		return
	}
	render.JSON(w, r, depositResponse)
}

// DepositAddressCtx is a function that returns a context with a depositResponse struct.
func DepositAddressCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var depositRequest depositResponse
		depositRequest.Code = currency.NewCode(strings.ToUpper(chi.URLParam(request, "asset")))
		exchangeNameReq := chi.URLParam(request, "exchange")
		chain := chi.URLParam(request, "chain")

		d := singleton.GetDealer()

		e, err := d.GetExchangeByName(exchangeNameReq)
		if err != nil {
			logrus.Errorf("failed to get exchange: %s\n", exchangeNameReq)
			render.Render(w, request, ErrInvalidRequest(err))
			return
		}

		pairs, _ := e.GetAvailablePairs(asset.Spot)
		for _, p := range pairs {
			logrus.Printf("pairs: %s\n", p.Quote.String())
			logrus.Printf("%s\n", p.Base.String())
		}

		subAccount, err := GetSubAccountByID(e, "")

		depositRequest.Chains, err = e.GetAvailableTransferChains(context.Background(), depositRequest.Code)
		logrus.Info(depositRequest.Chains)
		depositRequest.Asset = depositRequest.Code.Item
		depositRequest.AccountID = subAccount.ID

		// need to figure out chain selection
		// USDT FTX: [erc20 trx sol]
		// USDT Binance: [BNB BSC ETH SOL TRX]
		// USDT BTSE: []
		// USDT Bitfinex: [TETHERUSDTALG TETHERUSX TETHERUSDTBCH TETHERUSDTDVF TETHERUSO TETHERUSDTSOL TETHERUSDTHEZ TETHERUSE TETHERUSL TETHERUSS TETHERUSDTOMG]
		// USDT Kraken: [Tether USD (ERC20) Tether USD (TRC20)]
		// USDT Huobi:  [algousdt hrc20usdt solusdt trc20usdt usdt usdterc20]
		if e.GetName() == "Binance" {
			if chain == "erc20" {
				chain = "eth"
			}
		}

		if e.GetName() == "Huobi" {
			if chain == "trx" {
				chain = "trc20usdt"
			}
		}

		if e.GetName() == "Kraken" {
			if chain == "trx" {
				chain = "Tether USD (TRC20)"
			}
		}

		if e.GetName() == "BTSE" {
			chain = ""
		}

		if e.GetName() == "Bittrex" {
			chain = ""
		}

		if chain == "default" {
			if len(depositRequest.Chains) > 0 {
				chain = depositRequest.Chains[0]
			} else {
				chain = ""
			}
		}

		depositRequest.Address, err = e.GetDepositAddress(context.Background(), depositRequest.Code, depositRequest.AccountID, chain)
		if err != nil {
			logrus.Errorf("failed to get deposit address: %s\n", err)
			render.Render(w, request, ErrInvalidRequest(err))
			return
		}

		h, err := dealer.Holdings(d, e.GetName())
		if err != nil {
			logrus.Errorf("failed to get holdings: %s\n", err)
		}

		balance, err := h.CurrencyBalance(depositRequest.AccountID, asset.Spot, depositRequest.Code)
		if err != nil {
			logrus.Errorf("failed to get balance: %s\n", err)
		}

		depositRequest.Balance = balance.TotalValue

		depositRequest.Price, err = getDollarValue(e, depositRequest.Code, asset.Spot)
		if err != nil {
			logrus.Errorf("failed to get dollar value: %s\n", err)
		}

		ctx := context.WithValue(request.Context(), "response", &depositRequest)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

// GetSubAccountByID is a function that returns a subaccount by ID.
func GetSubAccountByID(e exchange.IBotExchange, accountId string) (account.SubAccount, error) {
	accounts, err := e.UpdateAccountInfo(context.Background(), asset.Spot)
	if err != nil {
		logrus.Errorf("failed to get exchange account: %s\n", err)
	}

	// return the first account if there's no other accounts
	if len(accounts.Accounts) == 1 {
		return accounts.Accounts[0], nil
	}

	for _, a := range accounts.Accounts {
		// return the main account for FTX
		if a.ID == "main" && e.GetName() == "FTX" {
			return a, nil
		}

		if a.ID == accountId {
			return a, nil
		}
	}
	return account.SubAccount{}, err
}

// TODO : needs refactoring and this can be done in a better way
// check first if with a loop for USDT, USD, BTC and ETH Pairs
// When found, start fetching price to get a dollar value

// getDollarValue returns the dollar value of the currency and route if there's no USDT pair available
func getDollarValue(e exchange.IBotExchange, code currency.Code, assetType asset.Item) (float64, error) {

	if code.Match(currency.USDT) || code.Match(currency.USD) || code.Match(currency.BUSD) || code.Match(currency.UST) {
		logrus.Infof("stable coin price is pegged to $1: %s\n", code.String())
		return 1, nil
	}

	//// get available pairs for spot
	//availablePairs, err := e.GetAvailablePairs(asset.Spot)
	//
	//if err != nil {
	//	logrus.Errorf("failed to get available pairs: %s\n", err)
	//}

	//tradeAblePairs := currency.Pairs{}
	//
	////INFO[0013] found pair: ETH/BRZ
	////INFO[0013] found pair: ETH/BTC
	////INFO[0013] found pair: ETH/EUR
	////INFO[0013] found pair: ETH/USD
	////INFO[0013] found pair: ETH/USDT
	//for _, p := range availablePairs {
	//	if p.ContainsCurrency(code) {
	//		tradeAblePairs = append(tradeAblePairs, p)
	//		logrus.Printf("found pair: %s\n", p.String())
	//	}
	//}

	tradeAblePairs := algo.MatchPairsForCurrency(e, code, asset.Spot)

	for _, p := range tradeAblePairs {
		if p.Quote.Match(currency.USD) {
			logrus.Printf("found USD pair: %s\n", p.String())
			ticker, err := e.FetchTicker(context.Background(), p, assetType)
			if err != nil {
				logrus.Errorf("failed to get price: %s\n", err)
			}
			return ticker.Last, nil
		}

		if p.Quote.Match(currency.USDT) {
			logrus.Printf("found USDT pair: %s\n", p.String())
			ticker, err := e.FetchTicker(context.Background(), p, assetType)
			if err != nil {
				logrus.Errorf("failed to get price: %s\n", err)
			}
			return ticker.Last, nil
		}
	}

	// TODO need to figure out the fastest pair routing for USD price when there's no USD or USDT as quote currency, code below sucks.

	//// Try to match with a BTC pair
	//p := currency.NewPair(code, currency.BTC) // ie VIA-BTC
	//if availablePairs.Contains(p, true) {     // confirm there's a BTC pair
	//	// if no USD pair is found, try BTC
	//	BTCUSDT := currency.NewPair(currency.BTC, currency.USDT)
	//	btcTicker, err := e.FetchTicker(context.Background(), BTCUSDT, assetType)
	//	if err != nil {
	//		logrus.Errorf("failed to get ticker: %s\n", err)
	//	}
	//
	//	ticker, err := e.FetchTicker(context.Background(), p, assetType) // get the ticker for the BTC pair (ie VIA-BTC)
	//	if err == nil {
	//		return ticker.Last * btcTicker.Last, nil // ie return VIABTC price * BTCUSDT price
	//	}
	//}

	return 0, errors.New("no USD, USDT or BTC pair found")
}

// TODO need to figure out the fastest pair routing for USD price when there's no USD or USDT as quote currency, code below sucks.

func GetDollarValueBTCPair(e exchange.IBotExchange, code currency.Code, assetType asset.Item) (float64, error) {
	p := currency.NewPair(code, currency.BTC) // ie VIA-BTC
	BTCUSDT := currency.NewPair(currency.BTC, currency.USDT)
	btcTicker, err := e.FetchTicker(context.Background(), BTCUSDT, assetType)
	if err != nil {
		return 0, err
	}

	ticker, err := e.FetchTicker(context.Background(), p, assetType) // get the ticker for the BTC pair (ie VIA-BTC)
	if err != nil {
		return 0, err
	}
	return ticker.Last * btcTicker.Last, err
}

func GetDollarValueETHPair(e exchange.IBotExchange, code currency.Code, assetType asset.Item) (float64, error) {
	p := currency.NewPair(code, currency.ETH) // ie VIA-BTC
	ETHUSDT := currency.NewPair(currency.ETH, currency.USDT)
	btcTicker, err := e.FetchTicker(context.Background(), ETHUSDT, assetType)
	if err != nil {
		return 0, err
	}

	ticker, err := e.FetchTicker(context.Background(), p, assetType) // get the ticker for the BTC pair (ie VIA-BTC)
	if err != nil {
		return 0, err
	}
	return ticker.Last * btcTicker.Last, err
}

func GetDollarValueBNBPair(e exchange.IBotExchange, code currency.Code, assetType asset.Item) (float64, error) {
	p := currency.NewPair(code, currency.BNB) // ie VIA-BTC
	BNBUSDT := currency.NewPair(currency.BNB, currency.USDT)
	btcTicker, err := e.FetchTicker(context.Background(), BNBUSDT, assetType)
	if err != nil {
		return 0, err
	}

	ticker, err := e.FetchTicker(context.Background(), p, assetType) // get the ticker for the BTC pair (ie VIA-BTC)
	if err != nil {
		return 0, err
	}
	return ticker.Last * btcTicker.Last, err
}

func GetDollarValueBUSDPair(e exchange.IBotExchange, code currency.Code, assetType asset.Item) (float64, error) {
	p := currency.NewPair(code, currency.BUSD)                       // ie VIA-BTC
	ticker, err := e.FetchTicker(context.Background(), p, assetType) // get the ticker for the BTC pair (ie VIA-BTC)
	if err != nil {
		return 0, err
	}
	return ticker.Last, err
}
