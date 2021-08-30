package webserver

import (
	"context"
	_ "fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/dealer"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
	"gopkg.in/errgo.v2/fmt/errors"
	"net/http"
	"strconv"
	_ "strconv"
	"strings"
)

type exchangeWithdrawResponse struct {
	//exchangeResponse withdraw.ExchangeResponse
	exchangeResponse withdraw.Response
}

func (e exchangeWithdrawResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "withdraw.html", nil)
	if err != nil {
		logrus.Errorf("error template: %s\n", err)
	}
}

func getExchangeWithdrawResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	exchangeResponse, ok := ctx.Value("response").(*exchangeWithdrawResponse) // TODO fix
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		render.Render(w, r, ErrWithdawRender(errors.Newf("Failed to renderWithdrawResponse")))
		return
	}
	render.Render(w, r, exchangeResponse)
	return
}

func WithdrawCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var err error
		assetInfo := new(Asset)

		dealer, err := dealer.New(engine.Settings{})
		if err != nil {
			logrus.Errorf("failed to create dealer %s\n", err)
		}

		exchangeNameReq := chi.URLParam(request, "exchange")
		destinationAddress := chi.URLParam(request, "destinationAddress")
		sizeReq := chi.URLParam(request, "size")

		size, err := strconv.ParseFloat(sizeReq, 64)
		if err != nil {
			logrus.Errorf("failed to parse size %s\n", err) // 3.14159265
			render.Render(w, request, ErrWithdawRender(err))
		}

		assetInfo.Code = currency.NewCode(strings.ToUpper(chi.URLParam(request, "asset")))
		assetInfo.Code.Item.Role = currency.Cryptocurrency

		exchangeEngine, err := dealer.ExchangeManager.GetExchangeByName(exchangeNameReq)
		if err != nil {
			logrus.Errorf("Failed to return exchange %s\n", err)
		}

		wi := &withdraw.Request{
			Exchange: exchangeEngine.GetName(),
			Currency: assetInfo.Code,
			Amount:   size,
			Type:     withdraw.Crypto,
			Crypto: withdraw.CryptoRequest{
				Address:    destinationAddress,
				AddressTag: "",
				FeeAmount:  0,
			},
		}

		response, err := dealer.WithdrawManager.SubmitWithdrawal(wi) //engine.Bot.WithdrawManager.SubmitWithdrawal(wi)
		if err != nil {
			logrus.Errorf("failed to withdraw crypto asset %s %s\n", assetInfo.Code, err)
			render.Render(w, request, ErrWithdawRender(err))
		}

		//response, err := exchangeEngine.WithdrawCryptocurrencyFunds(wi)
		//if err != nil {
		//	logrus.Errorf("failed to withdraw crypto asset %s %s\n", assetInfo.Code, err)
		//	render.Render(w, request, ErrWithdawRender(err))
		//}

		logrus.Infof("exchange withdraw response %v", response)
		ctx := context.WithValue(request.Context(), "response", exchangeWithdrawResponse{exchangeResponse: *response})
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
