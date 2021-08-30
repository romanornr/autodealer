// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package webserver

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/dealer"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine"
	"net/http"
	"strings"
	"time"
)

// DepositResponse which embeds an Asset. It defines an Asset struct and embeds the base asset struct
// which ensures that DepositResponse has access to the id, amount, and address fields
// It then specifies the Time field and specifies that it's time.Time
// In the end, DepositResponse has all the fields of both Asset and base Asset
type DepositResponse struct {
	*Asset
	Time time.Time `json:"time"`
}

// Render DepositResponse takes an interface as input (which means that we can use concrete types to render its BSON form).
// Look at the data type (a DepositResponse) and try to map its elements to BSON values. It's easy because they have identical names.
// Set the Timestamp field to the current time if it's empty.
// Return the input parameter, encoded as BSON.
func (d DepositResponse) Render(w http.ResponseWriter, r *http.Request) error {
	d.Time = time.Now()
	return nil
}

// DepositHandler is calling the ExecuteTemplate method with the first argument a http.ResponseWriter.
// The second argument will be the file named deposit.html inside the folder templates.
// The function can now be used as part of the router by adding the path to the function.
func DepositHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "deposit.html", nil)
	if err != nil {
		logrus.Errorf("error template: %s\n", err)
	}
}

// getDepositAddress extracts the `*Asset` from the context. This will make sure we are
// always working with the correct type and will allow us to return an error if the type is wrong Next, we check if the depositInfo is not nil.
// We will have to check this in order to ensure that problems in the middleware doesn't cause the whole request to fail.
// Our server will return a 422 error instead. Finally, we return the depositInfo variable to the user
func getDepositAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	assetInfo, ok := ctx.Value("depositInfo").(*Asset)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		render.Render(w, r, ErrNotFound)
		return
	}
	depositInfo := DepositResponse{assetInfo, time.Now()}
	render.Render(w, r, depositInfo)
	return
}

// DepositAddressCtx wraps the incoming API http request to request context and adds a depositInfo structure to it with the exchange and code assigned.
// This depositInfo structure is then used to look up the deposit address and deposit instructions for a particular exchange and asset pair.
// Next, it runs the next middleware handler in the chain. In our case, this is the router object, and this continues with the original request.
// The next handler is provided with the updated context and proceeds in the usual way.
func DepositAddressCtx(next http.Handler) http.Handler {
	assetInfo := new(Asset)
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		exchangeNameReq := chi.URLParam(request, "exchange")
		var settings engine.Settings
		d, err := dealer.New(settings)

		engineExchange, err := d.ExchangeManager.GetExchangeByName(exchangeNameReq)
		if err != nil {
			logrus.Errorf("failed to get exchange: %s\n", exchangeNameReq)
		}

		request = request.WithContext(context.WithValue(request.Context(), "exchange", engineExchange))
		logrus.Infof("request: %v\n", request)

		assetInfo.Exchange = engineExchange.GetName()
		assetInfo.Code = currency.NewCode(strings.ToUpper(chi.URLParam(request, "asset")))

		assetInfo.Address, err = engineExchange.GetDepositAddress(assetInfo.Code, "")
		if err != nil {
			logrus.Errorf("failed to get address: %s\n", err)
			render.Render(w, request, ErrInvalidRequest(err))
			return
		}

		ctx := context.WithValue(request.Context(), "depositInfo", assetInfo)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

//func getBalance(w http.ResponseWriter, r *http.Request) {
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
//}
//
//// BalanceCtx amend the request to be within the context of the asset name.
//// Get the account info from the exchange engine. Find the asset code in the account and get the balance.
//// Return the amended request. Add a balance cookie to the response.
//func BalanceCtx(next http.Handler) http.Handler {
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
//}
