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
	"gopkg.in/errgo.v2/fmt/errors"
)

// depositResponse struct with pre-defined members. This response may be populated by a deposit response from the exchange or an error.
// If it's populated by a deposit response from the exchange, the depositResponse struct fills the fields.
// If it's populated by an error, then the Err member will contain an error object.
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

// Render depositResponse assigns the Time value to a new special variable called time.Now().
// This value represents the time when the validation function is called.
// The d.Time value changes and gets updated with the current time every time we make a validation.
// Then it assigns the error value, which represents the errors data passed in the http response, to a new special variable called d.Err.
func (d *depositResponse) Render(w http.ResponseWriter, r *http.Request) error {
	d.Time = time.Now()
	return d.Err
}

// ErrDepositRender checks for the error. If the error exists, it then returns
// a DepositRender wrapper that contains that error in it.
func ErrDepositRender(err error) render.Renderer {
	return &depositResponse{
		Err: err,
	}
}

// DepositHandler is calling the ExecuteTemplate method with the first argument a http.ResponseWriter.
// The second argument will be the file named deposit.html inside the folder templates.
// The function can now be used as part of the router by adding the path to the function.
func DepositHandler(w http.ResponseWriter, _ *http.Request) {
	err := tpl.ExecuteTemplate(w, "deposit.html", nil)
	if err != nil {
		logrus.Errorf("error template: %s\n", err)
	}
}

// getDepositAddress extracts the `*DepositResponse` from the context. This will make sure we are
// always working with the correct type and will allow us to return an error if the type is wrong Next, we check if the depositInfo is not nil.
// We will have to check this in order to ensure that problems in the middleware doesn't cause the whole request to fail.
// Our server will return a 422 error instead. Finally, we return the depositInfo variable to the user
func getDepositAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	depositResponse, ok := ctx.Value("response").(*depositResponse)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.Render(w, r, ErrDepositRender(errors.Newf("Failed to render deposit response")))
		return
	}
	render.Render(w, r, depositResponse)
	return
}

// DepositAddressCtx wraps the incoming API http request to request context
// and adds a depositResponse structure to it with the exchange and code assigned.
// This depositResponse structure is then used to look up the deposit address
// and deposit instructions for a particular exchange and asset pair.
// Next, it runs the next middleware handler in the chain.
// In our case, this is the router object, and this continues with the original request.
// The next handler is provided with the updated context and proceeds in the usual way.
func DepositAddressCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		exchangeNameReq := chi.URLParam(request, "exchange")
		chain := chi.URLParam(request, "chain")
		accountId := make(chan string)

		d := GetDealerInstance()

		engineExchange, err := d.ExchangeManager.GetExchangeByName(exchangeNameReq)
		if err != nil {
			logrus.Errorf("failed to get exchange: %s\n", exchangeNameReq)
			render.Render(w, request, ErrInvalidRequest(err))
			return
		}

		go WithAccount(engineExchange, accountId)

		var depositRequest depositResponse
		depositRequest.Code = currency.NewCode(strings.ToUpper(chi.URLParam(request, "asset")))
		depositRequest.Chains, err = engineExchange.GetAvailableTransferChains(context.Background(), depositRequest.Code)
		logrus.Info(depositRequest.Chains)
		depositRequest.Asset = depositRequest.Code.Item
		depositRequest.Account = <-accountId

		if engineExchange.GetName() == "Binance" {
			if chain == "erc20" {
				chain = "ETH"
			}
			if depositRequest.Code == currency.LUNA && chain == "default" {
				chain = "LUNA"
			}
		}
		if chain == "default" {
			chain = ""
		}

		depositRequest.Address, err = engineExchange.GetDepositAddress(context.Background(), depositRequest.Code, depositRequest.Account, strings.ToLower(chain))
		if err != nil {
			logrus.Errorf("failed to get address: %s\n", err)
			render.Render(w, request, ErrInvalidRequest(err))
			return
		}

		ctx := context.WithValue(request.Context(), "response", &depositRequest)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

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
