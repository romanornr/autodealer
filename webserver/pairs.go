package webserver

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"gopkg.in/errgo.v2/fmt/errors"
	"net/http"
)

type pair struct {
	Name      string     `json:"name"`
	AssetType asset.Item `json:"assetType"`
}

type pairResponse struct {
	Pair []pair `json:"pair"`
}

// Render Pairs renders the pairs
func (p pairResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// FetchPairsCtx fetches pairs from the exchange
func FetchPairsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		d := GetDealerInstance()
		e, err := d.ExchangeManager.GetExchangeByName(chi.URLParam(request, "exchange"))
		if err != nil {
			logrus.Errorf("Failed to get exchange: %s\n", err)
		}
		assetTypes := e.GetAssetTypes(true)
		response := new(pairResponse)

		//for _, a := range assetTypes {
		//	pairs, err := e.FetchTradablePairs(context.Background(), a)
		//	if err != nil {
		//		continue
		//	}
		//	for _, p := range pairs {

		for _, a := range assetTypes {

			c, err := e.GetAvailablePairs(a)
			if err != nil {
				logrus.Errorf("Failed to get enabled pairs: %s\n", err)
			}

			formattedPair := c.Format("-", "", true)
			for _, p := range formattedPair {
				response.Pair = append(response.Pair, pair{Name: p.String(), AssetType: a})
			}
			//response.Pair = append(response.Pair, pair{Name: formattedPair, AssetType: a})
			//response.Pair = append(response.Pair, pair{Name: p, AssetType: a})
		}

		request = request.WithContext(context.WithValue(request.Context(), "response", response))
		next.ServeHTTP(w, request)
	})
}

// getPairsResponse is the response for the get pairs endpoint
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
