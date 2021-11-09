package webserver

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/transfer"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"gopkg.in/errgo.v2/fmt/errors"
	"net/http"
	"strconv"
	"time"
)

// TradeHandler handleHome is the handler for the '/trade' page request.
func TradeHandler(w http.ResponseWriter, r *http.Request) {
	if err := tpl.ExecuteTemplate(w, "trade.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Errorf("error template: %s\n", err)
		return
	}
}

// OrderResponse is the response for the '/order' request.
type OrderResponse struct {
	Response  order.SubmitResponse `json:"response"`
	Order     order.Submit         `json:"order"`
	Pair      string               `json:"pair"`
	QtyUSD    float64              `json:"qtyUSD"`
	Qty       float64              `json:"qty"`
	Price     float64              `json:"price"`
	Timestamp time.Time            `json:"timestamp"`
}

// Render Pairs renders the pairs
func (o OrderResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// getTradeResponse returns the trade response
func getTradeResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response, ok := ctx.Value("response").(*OrderResponse) // TODO fix
	if !ok {
		logrus.Errorf("Got unexpected response %T\n", response)
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.Render(w, r, transfer.ErrWithdawRender(errors.Newf("Failed to renderWithdrawResponse")))
		return
	}
	render.Render(w, r, response)
	return
}

// TradeCtx Handler handleHome is the handler for the '/trade' page request.
func TradeCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		exchangeNameReq := chi.URLParam(request, "exchange")
		p, err := currency.NewPairFromString(chi.URLParam(request, "pair"))
		if err != nil {
			logrus.Errorf("failed to parse pair: %s\n", chi.URLParam(request, "pair"))
		}

		qtyUSD, err := strconv.ParseFloat(chi.URLParam(request, "qty"), 32)
		if err != nil {
			logrus.Errorf("failed to parse qty %s\n", err)
		}

		assetItem, err := asset.New(chi.URLParam(request, "assetType"))
		if err != nil {
			logrus.Errorf("failed to parse assetType %s\n", err)
		}

		var orderType order.Type
		var side order.Side
		var postOnly bool

		switch chi.URLParam(request, "orderType") {
		case "marketBuy":
			orderType = order.Market
			side = order.Bid
		case "limitBuy":
			orderType = order.Limit
			side = order.Ask
			postOnly = true
		}

		d := GetDealerInstance()
		e, err := d.ExchangeManager.GetExchangeByName(exchangeNameReq)
		if err != nil {
			logrus.Errorf("failed to get exchange: %s\n", exchangeNameReq)
			return
		}

		// try to find out how to enable all pairs??
		d.Settings.EnableAllPairs = true
		d.Settings.EnableCurrencyStateManager = true

		price, err := e.UpdateTicker(context.Background(), p, assetItem)
		if err != nil {
			logrus.Errorf("failed to update ticker %s\n", err)
		}

		qty := qtyUSD / price.Ask
		logrus.Printf("%s quantity %f\n", p.String(), qty)

		o := order.Submit{
			ImmediateOrCancel: false,
			HiddenOrder:       false,
			FillOrKill:        false,
			PostOnly:          postOnly,
			ReduceOnly:        false,
			Leverage:          0,
			Price:             price.Ask,
			Amount:            qty,
			StopPrice:         0,
			LimitPriceUpper:   0,
			LimitPriceLower:   0,
			TriggerPrice:      0,
			TargetAmount:      0,
			ExecutedAmount:    0,
			RemainingAmount:   0,
			Fee:               0,
			Exchange:          e.GetName(),
			InternalOrderID:   "",
			ID:                "",
			AccountID:         "",
			ClientID:          "",
			ClientOrderID:     "",
			WalletAddress:     "",
			Offset:            "",
			Type:              orderType,
			Side:              side,
			Status:            "",
			AssetType:         assetItem,
			Date:              time.Now(),
			LastUpdated:       time.Time{},
			Pair:              p,
			Trades:            nil,
		}

		if err = o.Validate(); err != nil {
			logrus.Errorf("failed to validate order: %s\n", err)
		}

		submitResponse, err := d.SubmitOrderUD(context.Background(), e.GetName(), o, nil)//e.SubmitOrder(context.Background(), &o)
		if err != nil {
			logrus.Errorf("submit order failed: %s\n", err)
		}
		logrus.Printf("order response ID %s placed %t", submitResponse.OrderID, submitResponse.IsOrderPlaced)

		response := OrderResponse{
			Response:  submitResponse,
			Order:     o,
			Pair:      p.String(),
			QtyUSD:    qtyUSD,
			Qty:       qty,
			Price:     price.Ask,
			Timestamp: time.Now(),
		}

		ctx := context.WithValue(request.Context(), "response", &response)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
