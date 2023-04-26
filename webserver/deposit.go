// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package webserver

import (
	"context"
	"github.com/romanornr/autodealer/algo/shortestPath"
	"net/http"
	"strings"
	"time"

	"github.com/romanornr/autodealer/dealer"
	"github.com/romanornr/autodealer/singleton"
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
		chainReq := chi.URLParam(request, "chain")

		d := singleton.GetDealer()

		e, err := d.GetExchangeByName(exchangeNameReq)
		if err != nil {
			logrus.Errorf("failed to get exchange: %s\n", exchangeNameReq)
			render.Render(w, request, ErrInvalidRequest(err))
			return
		}

		subAccount, err := GetSubAccountByID(e, "")

		availableTransferChains, err := e.GetAvailableTransferChains(context.Background(), depositRequest.Code)
		depositRequest.Chains = availableTransferChains
		logrus.Infof("deposit request chains %v", availableTransferChains)
		depositRequest.Asset = depositRequest.Code.Item
		depositRequest.AccountID = subAccount.ID

		selectedChain := chainSelection(e.GetName(), chainReq, depositRequest.Chains)

		depositRequest.Address, err = e.GetDepositAddress(context.Background(), depositRequest.Code, depositRequest.AccountID, selectedChain)
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

// getDollarValue returns the dollar value of the currency and route if there's no USDT pair available
func getDollarValue(e exchange.IBotExchange, code currency.Code, assetType asset.Item) (float64, error) {

	if code.Item.Symbol == "USDT" || code.Item.Symbol == "USD" || code.Item.Symbol == "BUSD" || code.Item.Symbol == "UST" {
		return 1, nil
	}

	price, err := shortestPath.GetPrice(e, code, currency.USDT, assetType)
	if err != nil {
		logrus.Errorf("failed to get price: %s\n", err)
	}

	logrus.Printf("price: %f\n", price)

	return price, nil
}
