// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package webserver

import (
	"context"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/deposit"
)

// depositResponse is the response payload for deposit requests
type depositResponse struct {
	Asset   *currency.Item   `json:"asset"`
	Code    currency.Code    `json:"code"`
	Chains  []string         `json:"chains"`
	Address *deposit.Address `json:"address"`
	Time    time.Time        `json:"time"`
	Balance float64          `json:"balance"`
	Err     error            `json:"error"`
	Account string           `json:"account"`
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
		accountId := make(chan string)

		d := GetDealerInstance()

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

		go WithAccount(e, accountId)

		depositRequest.Chains, err = e.GetAvailableTransferChains(context.Background(), depositRequest.Code)
		logrus.Info(depositRequest.Chains)
		depositRequest.Asset = depositRequest.Code.Item
		depositRequest.Account = <-accountId

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

		if chain == "default" {
			chain = depositRequest.Chains[0]
		}

		depositRequest.Address, err = e.GetDepositAddress(context.Background(), depositRequest.Code, depositRequest.Account, chain)
		if err != nil {
			logrus.Errorf("failed to get address: %s\n", err)
			render.Render(w, request, ErrInvalidRequest(err))
			return
		}

		ctx := context.WithValue(request.Context(), "response", &depositRequest)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

// WithAccount returns a channel with the account id.
func WithAccount(e exchange.IBotExchange, accountId chan string) {
	accounts, err := e.FetchAccountInfo(context.Background(), asset.Spot)
	if err != nil {
		logrus.Errorf("failed to get exchange account: %s\n", err)
	}
	for _, a := range accounts.Accounts {
		accountId <- a.ID
		if a.ID == "main" {
			accountId <- "main"
			break
		}
	}
}

// func getBalance(w http.ResponseWriter, r *http.Request) {
//	exchangeName := r.Context().Value("exchange").(exchange.IBotExchange)
//	//account := r.Context().Value("accounts").(*exchange.AccountInfo)
//	balance := r.Context().Value("balance").(float64)
//
//	res := Asset{
//		Exchange: exchangeName.GetName(),
//		Code:     r.Context().Value("assetInfo").(Asset).Code,
//		Address:  r.Context().Value("assetInfo").(Asset).Address,
//		Balance:  strconv.FormatFloat(balance, 'f', -1, 64),
//	}
//
//	logrus.Infof("res %+v", res)
//	render.JSON(w, r, res)
// }
//
// // BalanceCtx amend the request to be within the context of the asset name.
// // Get the account info from the exchange engine. Find the asset code in the account and get the balance.
// // Return the amended request. Add a balance cookie to the response.
// func BalanceCtx(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
//		assetInfo := new(Asset)
//		exchangeEngine := request.Context().Value("exchange").(exchange.IBotExchange)
//		//accountID := request.Context().Value("accountID").(string)
//		assetInfo.Code = currency.NewCode(strings.ToUpper(chi.URLParam(request, "asset")))
//
//		accounts, err := exchangeEngine.FetchAccountInfo(asset.Spot)
//		if err != nil {
//			logrus.Errorf("failed to fetch accounts: %s\n", err)
//		}
//
//		account := accounts.Accounts[0]
//		for _, c := range account.Currencies {
//			if c.CurrencyName == assetInfo.Code {
//				logrus.Info(c.TotalValue)
//				assetInfo.Balance = fmt.Sprintf("%f", c.TotalValue)
//			}
//		}
//
//		logrus.Infof("balance: %s\n", assetInfo.Balance)
//		request = request.WithContext(context.WithValue(request.Context(), "balance", assetInfo.Balance))
//
//		next.ServeHTTP(w, request)
//	})
// }
