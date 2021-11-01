package webserver

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"gopkg.in/errgo.v2/fmt/errors"
	"net/http"
)

// pairResponse is the response for the pair endpoint
type pairResponse struct {
	Pairs []string `json:"pairs"`
}

// Render Pairs renders the pairs
func (p pairResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// getPairsResponse renders the pairs response
func getPairsResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response, ok := ctx.Value("response").(*pairResponse)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.Render(w, r, ErrDepositRender(errors.Newf("Failed to render deposit response")))
		return
	}
	render.Render(w, r, response)
	return
}

// FetchPairsCtx is a middleware that fetches pairs from the exchange
func FetchPairsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		d := GetDealerInstance()
		assetItem, err := asset.New(chi.URLParam(request, "assetItem"))

		e, err := d.ExchangeManager.GetExchangeByName(chi.URLParam(request, "exchange"))
		pairs, err := e.FetchTradablePairs(context.Background(), assetItem)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

		request = request.WithContext(context.WithValue(request.Context(), "response", &pairResponse{Pairs: pairs}))
		next.ServeHTTP(w, request)
    })
}
