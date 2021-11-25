package webserver

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

type pair struct {
	Name      string     `json:"name"`
	AssetType asset.Item `json:"assetType"`
}

type pairResponse struct {
	Pair []pair `json:"pair"`
}

// FetchPairsCtx fetches pairs from the exchange
func FetchPairsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		d := GetDealerInstance()
		e, err := d.ExchangeManager.GetExchangeByName(chi.URLParam(request, "exchange"))
		if err != nil {
			logrus.Errorf("Failed to get exchange: %s\n", err)
		}

		/// enable for Bittrex
		if e.GetName() == "Bittrex" || e.GetName() == "Huobi" {
			if err := e.GetBase().CurrencyPairs.SetAssetEnabled(asset.Spot, true); err != nil {
				logrus.Errorf("Failed to enable asset: %s\n", err)
			}
		}

		assetTypes := e.GetAssetTypes(true)
		response := new(pairResponse)

		for _, a := range assetTypes {
			c, err := e.GetAvailablePairs(a)
			if err != nil {
				logrus.Errorf("Failed to get pairs: %s\n", err)
			}

			for _, p := range c {
				response.Pair = append(response.Pair, pair{
					Name:      p.Format("-", true).String(),
					AssetType: a,
				})
			}
		}

		//	formattedPair := c.Format("-", "", true)
		//	for _, p := range formattedPair {
		//		response.Pair = append(response.Pair, pair{Name: p.String(), AssetType: a})
		//	}
		//}

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
		render.Status(r, http.StatusUnprocessableEntity)
		return
	}
	render.JSON(w, r, response)
	return
}
