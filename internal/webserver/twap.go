package webserver

//import (
//	"context"
//	"net/http"
//	"strconv"
//	"strings"
//	"time"
//
//	"github.com/go-chi/chi/v5"
//	"github.com/go-chi/render"
//	"github.com/romanornr/autodealer/internal/algo"
//	"github.com/sirupsen/logrus"
//	"github.com/thrasher-corp/gocryptotrader/currency"
//	"github.com/thrasher-corp/gocryptotrader/engine"
//	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
//	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
//)
//
//func getTwap(w http.ResponseWriter, r *http.Request) {
//	ctx := r.Context()
//	val, ok := ctx.Value("twap").(*algo.TWAP)
//	if !ok {
//		http.Error(w, http.StatusText(422), 422)
//		render.Render(w, r, ErrNotFound)
//		return
//	}
//	render.Render(w, r, val)
//	return
//}
//
//func TwapCtx(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
//		baseCode := strings.ToUpper(chi.URLParam(request, "base"))
//		quoteCode := strings.ToUpper(chi.URLParam(request, "quote"))
//		quantityReq := chi.URLParam(request, "quantity")
//		exchangeEngine, _ := engine.Bot.GetExchangeByName(chi.URLParam(request, "exchange"))
//		quantity, err := strconv.ParseFloat(quantityReq, 64)
//		if err != nil {
//			logrus.Errorf("") // 3.14159265
//		}
//
//		pair := currency.NewPair(currency.NewCode(baseCode), currency.NewCode(quoteCode))
//		logrus.Infof("pair %s\n", pair)
//
//		wap := algo.TWAP{
//			Exchange:     exchangeEngine,
//			Pair:         pair,
//			Asset:        asset.Spot,
//			MaxChangePct: 10,
//			Start:        time.Now(),
//			End:          time.Now().Add(time.Hour * 24),
//			WapPrice:     0,
//			OverBought:   false,
//			Side:         order.Buy,
//		}
//
//		go wap.Execute(quantity, 3, 5, "aggressive")
//		ctx := context.WithValue(request.Context(), "twap", wap)
//		next.ServeHTTP(w, request.WithContext(ctx))
//	})
//}
