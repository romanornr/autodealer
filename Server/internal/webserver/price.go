package webserver

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/internal/algo/shortestPath"
	"github.com/romanornr/autodealer/internal/singleton"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"net/http"
	"strings"
)

// PriceResponse represents the response from the price endpoint
type PriceResponse struct {
	Exchange  string        `json:"exchange"`
	Base      currency.Code `json:"base"`
	Quote     currency.Code `json:"quote"`
	Price     float64       `json:"price"`
	AssetType asset.Item    `json:"type"`
	// Error handling
	Error string `json:"error,omitempty"` // TODO find out how to improve error handling for API response
}

// GetPrice returns the price of a given currency pair
func getPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response, ok := ctx.Value("response").(*PriceResponse)
	if !ok {
		logrus.Error("could not get response from context")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	render.JSON(w, r, response)
}

// PriceCtx is the context for the price endpoint
func PriceCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var response PriceResponse

		response.Base = currency.NewCode(strings.ToUpper(chi.URLParam(r, "base")))
		response.Quote = currency.NewCode(strings.ToUpper(chi.URLParam(r, "quote")))

		// check if base is valid
		if response.Base.IsEmpty() || response.Quote.IsEmpty() {
			response.Error = "invalid base or quote code"
			logrus.Error("invalid base or quote code")
			ctx := context.WithValue(r.Context(), "response", &response)
			next.ServeHTTP(w, r.WithContext(ctx))

			return
		}

		switch chi.URLParam(r, "assetType") {
		case "spot":
			response.AssetType = asset.Spot
		case "futures":
			response.AssetType = asset.Futures
		default:
			logrus.Printf("assetType: %s", r.URL.Query().Get("assetType"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		d := singleton.GetDealer()
		response.Exchange = chi.URLParam(r, "exchange")
		e, err := d.ExchangeManager.GetExchangeByName(response.Exchange)
		if err != nil {
			logrus.Errorf("Failed to get exchange: %s\n", err)
		}

		price, err := shortestPath.GetPrice(e, response.Base, response.Quote, response.AssetType)
		if err != nil {
			logrus.Errorf("Failed to get price: %s\n", err)
			response.Error = err.Error()
		}

		response.Price = price

		ctx := context.WithValue(r.Context(), "response", &response)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
