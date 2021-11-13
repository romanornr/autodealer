package webserver

import (
	"context"
	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/dealer"
	"github.com/sirupsen/logrus"
	"net/http"
)

// getHoldings Handler returns all holdings for a given exchange
func getHoldingsExchangeResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response, ok := ctx.Value("response").(*dealer.ExchangeHoldings)
	if !ok {
		logrus.Error("could not get response from context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := render.Render(w, r, response); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
	return

}

func HoldingsExchangeCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		d := GetDealerInstance()
		//holdings, err := dealer.Holdings(d, chi.URLParam(request, "exchange"))
		holdings, err := dealer.Holdings(d, "ftx")
		if err != nil {
			logrus.Errorf("Error getting holdings: %s\n", err)
		}

		logrus.Println(holdings)

		ctx := context.WithValue(request.Context(), "response", holdings)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
