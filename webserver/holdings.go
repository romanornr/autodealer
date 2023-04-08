package webserver

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/romanornr/autodealer/singleton"
	"github.com/romanornr/autodealer/subaccount"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/dealer"
	"github.com/sirupsen/logrus"
)

// getHoldings Handler returns all holdings for a given exchange
func getHoldingsExchangeResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response, ok := ctx.Value("response").(*dealer.CurrencyBalance)
	if !ok {
		logrus.Error("could not get response from context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, response)
}

// HoldingsExchangeCtx middleware adds the holdings to the context
func HoldingsExchangeCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		d := singleton.GetDealer()
		e, err := d.GetExchangeByName(chi.URLParam(request, "exchange"))
		if err != nil {
			logrus.Errorf("could not get exchange: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		code := currency.NewCode(strings.ToUpper(chi.URLParam(request, "asset")))

		holdings, err := dealer.Holdings(d, e.GetName())
		if err != nil {
			logrus.Errorf("Error getting holdings: %s\n", err)
		}

		subAccount, err := subaccount.GetByID(e, "")

		response, err := holdings.CurrencyBalance(subAccount.ID, asset.Spot, code)
		if err != nil {
			logrus.Errorf("Error getting currency balance: %s\n", err)
		}

		ctx := context.WithValue(request.Context(), "response", &response)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
