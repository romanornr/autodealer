package webserver

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"gopkg.in/errgo.v2/fmt/errors"
	"net/http"
)

// example response
//{
//   "pair":[
//      [
//         "futures",
//         "1INCH-PERP"
//      ],
//      [
//         "spot",
//         "1INCH-USD"
//      ]
//   ]
//}

// pairResponse is the response for the pair endpoint
type pairResponse struct {
	Pairs [][]string `json:"pair"`
}

// Render Pairs renders the pairs
func (p pairResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// FetchPairsCtx fetches the pairs from the exchange
func FetchPairsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		d := GetDealerInstance()
		e, err := d.ExchangeManager.GetExchangeByName(chi.URLParam(request, "exchange"))
		if err != nil {
			logrus.Errorf("Failed to get exchange: %s\n", err)
		}
		types := e.GetAssetTypes(true)

		response := new(pairResponse)

		for _, x := range types {
			pairs, err := e.FetchTradablePairs(context.Background(), x)
			if err != nil {
				continue
			}
			for _, p := range pairs {
				response.Pairs = append(response.Pairs, []string{x.String(), p})
			}
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
