package webserver

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"net/http"
	"strconv"
	"strings"
)

func getBalance(w http.ResponseWriter, r *http.Request) {
	exchangeName := r.Context().Value("exchange").(exchange.IBotExchange)
	//account := r.Context().Value("accounts").(*exchange.AccountInfo)
	balance := r.Context().Value("balance").(float64)

	res := Asset{
		Exchange: exchangeName.GetName(),
		Code:     r.Context().Value("assetInfo").(Asset).Code,
		Address:  r.Context().Value("assetInfo").(Asset).Address,
		Balance:  strconv.FormatFloat(balance, 'f', -1, 64),
	}

	logrus.Infof("res %+v", res)
	render.JSON(w, r, res)
}

// BalanceCtx amend the request to be within the context of the asset name.
// Get the account info from the exchange engine. Find the asset code in the account and get the balance.
// Return the amended request. Add a balance cookie to the response.
func BalanceCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		assetInfo := new(Asset)
		exchangeEngine := request.Context().Value("exchange").(exchange.IBotExchange)
		//accountID := request.Context().Value("accountID").(string)
		assetInfo.Code = currency.NewCode(strings.ToUpper(chi.URLParam(request, "asset")))

		accounts, err := exchangeEngine.FetchAccountInfo(asset.Spot)
		if err != nil {
			logrus.Errorf("failed to fetch accounts: %s\n", err)
		}

		account := accounts.Accounts[0]
		for _, c := range account.Currencies {
			if c.CurrencyName == assetInfo.Code {
				logrus.Info(c.TotalValue)
				assetInfo.Balance = fmt.Sprintf("%f", c.TotalValue)
			}
		}

		logrus.Infof("balance: %s\n", assetInfo.Balance)
		request = request.WithContext(context.WithValue(request.Context(), "balance", assetInfo.Balance))

		next.ServeHTTP(w, request)
	})
}
