package webserver

import (
	"context"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"net/http"

	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/transfer"
	"github.com/sirupsen/logrus"
)

func bankTransferHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "bank.html", nil)
	if err != nil {
		logrus.Errorf("error template: %s\n", err)
	}
}

func getBankTransfer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	exchangeResponse, ok := ctx.Value("response").(*transfer.ExchangeWithdrawResponse)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.JSON(w, r, http.StatusUnprocessableEntity)
		return
	}

	render.JSON(w, r, exchangeResponse)
}

func BankTransferCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		//currencyCode := currency.NewCode(chi.URLParam(request, "exchange"))     // TODO fix currency.EUR
		currencyCode := currency.EUR

		submitResponse, err := transfer.KrakenConvertUSDT(currencyCode)
		if err != nil {
			logrus.Errorf("Failed to sell USDT to Euro: %s\n", err)
			render.Status(request, http.StatusUnprocessableEntity)
			render.JSON(w, request, http.StatusUnprocessableEntity)
			return
		}
		logrus.Infof("submit response %v\n", submitResponse)

		response, err := transfer.KrakenInternationalBankAccountWithdrawal(currencyCode)
		if err != nil {
			logrus.Errorf("Failed to withdraw EUR from bank account: %s\n", err)
			render.Status(request, http.StatusUnprocessableEntity)
			render.JSON(w, request, http.StatusUnprocessableEntity)
			return
		}

		ctx := context.WithValue(request.Context(), "response", &response)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
