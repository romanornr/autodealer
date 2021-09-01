package webserver

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
)

func (history WithdrawHistoryResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type WithdrawHistoryResponse struct {
	History []exchange.WithdrawalHistory `json:"history"`
}

// get withdrawal history from exchange from an asset
func getWithdrawHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	withdrawHistory, ok := ctx.Value("history").([]exchange.WithdrawalHistory)
	if !ok {
		http.Error(w, http.StatusText(400), 400)
		render.Render(w, r, ErrNotFound)
		return
	}
	history := WithdrawHistoryResponse{
		History: withdrawHistory,
	}
	render.Render(w, r, history)
	return
}

// TODO fix FTX ERRO[0006] failed fetch history: not yet implemented
// only works for binance
func withdrawHistoryCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var ctx context.Context
		exchangeNameReq := chi.URLParam(request, "exchange")
		exchangeEngine, _ := engine.Bot.GetExchangeByName(exchangeNameReq)
		code := currency.NewCode(strings.ToUpper(chi.URLParam(request, "asset")))
		history, err := exchangeEngine.GetWithdrawalsHistory(code)
		if err != nil {
			logrus.Errorf("failed fetch history: %s", err)
		}
		ctx = context.WithValue(request.Context(), "history", history)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
