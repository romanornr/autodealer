package webserver

import (
	"context"
	"github.com/romanornr/autodealer/internal/singleton"
	"github.com/thrasher-corp/gocryptotrader/currency"
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

type exchangeAsset struct {
	Code string `json:"code"`
}

type exchangeAssetResponse struct {
	ExchangeAsset []exchangeAsset `json:"assets"`
}

// FetchPairsCtx fetches pairs from the exchange
func FetchPairsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		d := singleton.GetDealer()
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

		currencyPairFormat := currency.PairFormat{
			Uppercase: true,
			Delimiter: "-",
			Separator: "",
			Index:     "",
		}

		for _, a := range assetTypes {
			c, err := e.GetAvailablePairs(a)
			if err != nil {
				logrus.Errorf("Failed to get pairs: %s\n", err)
			}

			for _, p := range c {
				response.Pair = append(response.Pair, pair{
					Name:      currencyPairFormat.Format(p), //p.Format()//p.Format("-", true).String(),
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
}

// getAssetList is the response for the get currencies endpoint
func getAssetList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response, ok := ctx.Value("response").(*exchangeAssetResponse)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.Status(r, http.StatusUnprocessableEntity)
		return
	}
	render.JSON(w, r, response)
}

// AssetListCtx fetches currencies from the exchange
func AssetListCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {

		d := singleton.GetDealer()
		e, err := d.ExchangeManager.GetExchangeByName(chi.URLParam(request, "exchange"))
		if err != nil {
			logrus.Errorf("Failed to get exchange: %s\n", err)
		}

		pairs, err := e.GetAvailablePairs(asset.Spot)
		if err != nil {
			logrus.Errorf("Failed to get pairs: %s\n", err)
		}

		response := new(exchangeAssetResponse)

		// `m` is used to store unique assets for easy search in the loop below
		m := make(map[string]struct{})

		for i := 0; i < len(pairs); i++ {
			b := pairs[i].Base.String()
			q := pairs[i].Quote.String()

			// if there is no such currency in map, add it to map and assets
			if _, ok := m[b]; !ok {
				m[b] = struct{}{}
				response.ExchangeAsset = append(response.ExchangeAsset, exchangeAsset{Code: b})
			}

			if _, ok := m[q]; !ok {
				m[q] = struct{}{}
				response.ExchangeAsset = append(response.ExchangeAsset, exchangeAsset{Code: q})
			}
		}

		request = request.WithContext(context.WithValue(request.Context(), "response", response))
		next.ServeHTTP(w, request)
	})
}
