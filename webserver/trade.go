package webserver

import (
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"net/http"
)

// TradeHandler handleHome is the handler for the '/trade' page request.
func TradeHandler(w http.ResponseWriter, r *http.Request) {
	if err := tpl.ExecuteTemplate(w, "trade.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Errorf("error template: %s\n", err)
		return
	}
}

func getTradeResponse(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	//exchangeResponse, ok := ctx.Value("response").(*transfer.ExchangeWithdrawResponse) // TODO fix
	//if !ok {
	//	logrus.Errorf("Got unexpected response %T instead of *ExchangeWithdrawResponse", exchangeResponse)
	//	//http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
	//	//render.Render(w, r, transfer.ErrWithdawRender(errors.Newf("Failed to renderWithdrawResponse")))
	//	return
	//}
	//
	//render.Render(w, r, exchangeResponse)

	return
}

func TradeCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		exchangeNameReq := chi.URLParam(request, "exchange")

		d := GetDealerInstance()

		e, err := d.ExchangeManager.GetExchangeByName(exchangeNameReq)
		if err != nil {
			logrus.Errorf("failed to get exchange: %s\n", exchangeNameReq)
			return
		}

		e.SupportsAsset(asset.Spot)

		//e.SetPairs()
		//currency.NewPair()
		//
		//
		//order.Submit{
		//	ImmediateOrCancel: false,
		//	HiddenOrder:       false,
		//	FillOrKill:        false,
		//	PostOnly:          false,
		//	ReduceOnly:        false,
		//	Leverage:          0,
		//	Price:             0,
		//	Amount:            0,
		//	StopPrice:         0,
		//	LimitPriceUpper:   0,
		//	LimitPriceLower:   0,
		//	TriggerPrice:      0,
		//	TargetAmount:      0,
		//	ExecutedAmount:    0,
		//	RemainingAmount:   0,
		//	Fee:               0,
		//	Exchange:          "",
		//	InternalOrderID:   "",
		//	ID:                "",
		//	AccountID:         "",
		//	ClientID:          "",
		//	ClientOrderID:     "",
		//	WalletAddress:     "",
		//	Offset:            "",
		//	Type:              "",
		//	Side:              "",
		//	Status:            "",
		//	AssetType:         "",
		//	Date:              time.Time{},
		//	LastUpdated:       time.Time{},
		//	Pair:              currency.Pair{},
		//	Trades:            nil,
		//}

		//ctx := context.WithValue(request.Context(), "response", &depositRequest)
		//next.ServeHTTP(w, request.WithContext(ctx))

	})
}
