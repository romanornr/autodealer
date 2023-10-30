package webserver

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/singleton"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"net/http"
	"strings"
)

func getAvailableTransferChainsResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response, ok := ctx.Value("chains").(*[]string)
	if !ok {
		logrus.Errorf("failed to get available transfer chains response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, response)
}

func AvailableTransferChainsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		assetInfo := new(Asset)

		d, err := singleton.GetDealer(request.Context())
		if err != nil {
			logrus.Errorf("failed to get dealer: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		e, err := d.GetExchangeByName(chi.URLParam(request, "exchange"))
		if err != nil {
			logrus.Errorf("failed to get exchange %s by name: %s", e.GetName(), err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		assetInfo.Code = currency.NewCode(strings.ToUpper(chi.URLParam(request, "asset")))
		assetInfo.Exchange = e.GetName()

		fmt.Println(assetInfo.Code)

		chains, err := e.GetAvailableTransferChains(context.Background(), assetInfo.Code)
		if err != nil {
			logrus.Errorf("failed to get available transfer chains: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		request = request.WithContext(context.WithValue(request.Context(), "chains", &chains))
		next.ServeHTTP(w, request)
	})
}
