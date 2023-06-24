package webserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/romanornr/autodealer/algo/twap"
	"github.com/romanornr/autodealer/singleton"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/orderbuilder"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

// TradeHandler handleHome is the handler for the '/trade' page request.
func (h *Handler) TradeHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.tpl.ExecuteTemplate(w, "trade.html", nil); err != nil {
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

// getTradeResponse returns the trade response
func getTradeResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response, ok := ctx.Value("response").(*OrderResponse) // TODO fix
	if !ok {
		logrus.Errorf("Got unexpected response %T\n", response)
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.JSON(w, r, ErrRender(errors.New("failed to get trade response")))
		return
	}
	render.JSON(w, r, response)
}

// TradeCtx is the context for the '/trade' request.
// trade/{exchange}/{pair}/{qty}/{assetType}/{orderType}/{side}
func TradeCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		exchangeNameReq := chi.URLParam(request, "exchange")
		p, err := currency.NewPairFromString(chi.URLParam(request, "pair"))
		if err != nil {
			logrus.Errorf("failed to parse pair: %s\n", chi.URLParam(request, "pair"))
		}

		//assetItem := asset.Item(chi.URLParam(request, "assetType"))
		assetItem, err := asset.New(chi.URLParam(request, "assetType"))
		if err != nil {
			logrus.Errorf("failed to parse asset: %s\n", chi.URLParam(request, "assetType"))
		}

		fmt.Println(assetItem.IsValid())

		if !assetItem.IsValid() {
			logrus.Errorf("failed to parse assetType: %s\n", chi.URLParam(request, "assetType"))
		}

		side, err := order.StringToOrderSide(chi.URLParam(request, "side"))
		if err != nil {
			logrus.Errorf("failed to parse side: %s\n", chi.URLParam(request, "side"))
		}

		d := singleton.GetDealer()
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

		fmt.Println(price)

		qtyUSD, err := strconv.ParseFloat(chi.URLParam(request, "qty"), 64)
		if err != nil {
			logrus.Errorf("failed to parse qty %s\n", err)
		}

		orderType, err := order.StringToOrderType(chi.URLParam(request, "orderType"))
		if err != nil {
			logrus.Errorf("failed to parse orderType %s\n", err)
		}

		fmt.Printf("last price:%f\n", price.Last)

		//qty := qtyUSD / price.Last
		qty := 0.443
		//subAccount, err := GetSubAccountByID(e, "")

		ob := orderbuilder.NewOrderBuilder()
		ob.
			AtExchange(e.GetName()).
			//ForAccountID(subAccount.ID).
			ForCurrencyPair(p).
			WithAssetType(assetItem).
			ForPrice(price.Last).
			WithAmount(qty).
			UseOrderType(orderType).
			SetQuoteAmount(qtyUSD).
			SetSide(side)

		newOrder, err := ob.Build()

		logrus.Printf("new order: %+v\n", newOrder)

		logrus.Printf("%s quantity %f\n", p.String(), qty)

		newOrder.QuoteAmount = 15

		submitResponse, err := d.SubmitOrderUD(context.Background(), e.GetName(), *newOrder, nil)
		if err != nil {
			logrus.Errorf("submit order failed: %s\n", err)
		}
		logrus.Printf("order response ID %s placed %s", submitResponse.OrderID, submitResponse.Status.String())

		response := OrderResponse{
			Response:  *submitResponse,
			Order:     *newOrder,
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

// getTwapResponse returns the twap response
func getTwapResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response, ok := ctx.Value("response").(*twap.Payload) // TODO fix
	if !ok {
		logrus.Errorf("Got unexpected response %T\n", response)
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.JSON(w, r, ErrRender(errors.New("failed to get twap response")))
		return
	}
	render.JSON(w, r, response)
}

// TWAPCtx is the context for the '/twap' request
// twap/{exchange}/{pair}/{qty}/{assetType}/{orderType}/{side}
func TWAPCtx(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {

		exchangeNameReq := chi.URLParam(request, "exchange")
		p, err := currency.NewPairFromString(chi.URLParam(request, "pair"))
		if err != nil {
			logrus.Errorf("failed to parse pair: %s\n", chi.URLParam(request, "pair"))
		}

		//assetItem := asset.Item(chi.URLParam(request, "assetType"))
		assetItem, err := asset.New(chi.URLParam(request, "assetType"))
		if err != nil {
			logrus.Errorf("failed to parse asset: %s\n", chi.URLParam(request, "assetType"))
		}

		if !assetItem.IsValid() {
			logrus.Errorf("failed to parse assetType: %s\n", chi.URLParam(request, "assetType"))
		}

		side, err := order.StringToOrderSide(chi.URLParam(request, "side"))
		if err != nil {
			logrus.Errorf("failed to parse side: %s\n", chi.URLParam(request, "side"))
		}

		d := singleton.GetDealer()
		e, err := d.ExchangeManager.GetExchangeByName(exchangeNameReq)
		if err != nil {
			logrus.Errorf("failed to get exchange: %s\n", exchangeNameReq)
			return
		}

		hours := chi.URLParam(request, "hours")
		h, err := strconv.ParseFloat(hours, 32)
		if err != nil {
			logrus.Errorf("hours %f\n", h)
		}

		minutes := chi.URLParam(request, "minutes")
		m, err := strconv.ParseFloat(minutes, 32)
		if err != nil {
			logrus.Errorf("hours %f\n", h)
		}

		subAccount, err := GetSubAccountByID(e, "")

		qtyUSD, err := strconv.ParseFloat(chi.URLParam(request, "qty"), 64)
		if err != nil {
			logrus.Errorf("failed to parse qty %s\n", err)
		}

		orderType, err := order.StringToOrderType(chi.URLParam(request, "orderType"))
		if err != nil {
			logrus.Errorf("failed to parse orderType %s\n", err)
		}

		//price, err := e.UpdateTicker(context.Background(), p, assetItem)
		//if err != nil {
		//	logrus.Errorf("failed to update ticker %s\n", err)
		//}

		targetAmountQuote := qtyUSD

		logrus.Printf("%s targetAmountQuote %f\n", p.String(), targetAmountQuote)

		var orderPayload = twap.Payload{
			Exchange:          e.GetName(),
			AccountID:         subAccount.ID,
			Pair:              p,
			Asset:             assetItem,
			Start:             time.Now(),
			End:               time.Now().Add(time.Hour * time.Duration(h)).Add(time.Minute * time.Duration(m)),
			OrderType:         orderType,
			TargetAmountQuote: targetAmountQuote,
			Side:              side,
		}

		task, err := twap.NewTwapTask(orderPayload)
		if err != nil {
			logrus.Errorf("failed to create twap order task %s\n", err)
		}

		//algo.Execute(orderPayload, d)

		client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
		defer client.Close()

		//info, err := client.Enqueue(task)
		//if err != nil {
		//	log.Fatalf("could not enqueue task: %v", err)
		//}
		//
		//logrus.Printf("enqueued task: id=%s queue=%s", info.ID, info.Queue)

		info, err := client.Enqueue(task)
		if err != nil {
			log.Fatalf("could not enqueue task: %v", err)
		}

		logrus.Printf("enqueued task: id=%s\n", info.ID)

		response := orderPayload
		ctx := context.WithValue(request.Context(), "response", &response)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
