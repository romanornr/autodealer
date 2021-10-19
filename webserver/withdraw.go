package webserver

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/dealer"
	"github.com/romanornr/autodealer/transfer"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
	"gopkg.in/errgo.v2/fmt/errors"
)

// ExchangeWithdrawResponse is a struct that is designed to represent
// the response from the ExchangesWithdrawals API call.
// ExchangesWithdrawals is a simple function which returns deposit, withdraw,
// trade and withdrawal information so we will only add the information there which we are interested in:
// ExchangesResponse is a struct that includes information about the request as well as the response,
// we're only interested in the response hence why we've added resp.
// We've used the Withdrawal struct as that is the response from the exchange (withdraw.ExchangeResponse).
// The Exchange key is the exchange used to make the request.
// The Type key represents the type of information requested in the Call function.
// The DestinationAddress is the address the withdrawal was sent to if the request used the DepositAddress field.
// The Time key is when the request was made and the Err field returns errors if any occurred.
type ExchangeWithdrawResponse struct {
	ExchangeResponse   *withdraw.ExchangeResponse
	Exchange           string               `json:"exchange"`
	Type               withdraw.RequestType `json:"type"`
	DestinationAddress string               `json:"destination"`
	Time               time.Time            `json:"time"`
	Err                error                `json:"err"`
}

// WithdrawHandler is calling the ExecuteTemplate method with the first argument a http.ResponseWriter.
// The second argument will be the file named deposit.html inside the folder templates.
// The function can now be used as part of the router by adding the path to the function.
func WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "withdraw.html", nil)
	if err != nil {
		logrus.Errorf("error template: %s\n", err)
	}
}

// getExchangeWithdrawResponse is the reward for performing an exchange withdraw transaction. It's called
// as part of what is called an exchange event. The received json request is consistent with what is
// expected for what the function defines.
func getExchangeWithdrawResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	exchangeResponse, ok := ctx.Value("response").(*transfer.ExchangeWithdrawResponse) // TODO fix
	if !ok {
		logrus.Errorf("Got unexpected response %T instead of *ExchangeWithdrawResponse", exchangeResponse)
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.Render(w, r, transfer.ErrWithdawRender(errors.Newf("Failed to renderWithdrawResponse")))
		return
	}

	render.Render(w, r, exchangeResponse)

	return
}

// WithdrawCtx is an HTTP handler function which stores the request input with the help of chi.URLParams get method
// in the response and call the createExchangeWithdrawResponse to create an exchange withdrawal transaction
// for the specified exchange
func WithdrawCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var err error
		assetInfo := new(Asset)

		d, err := dealer.NewBuilder().Build()
		if err != nil {
			logrus.Errorf("failed to create a dealer %s\n", err)
		}

		exchangeNameReq := chi.URLParam(request, "exchange")
		destinationAddress := chi.URLParam(request, "destinationAddress")
		sizeReq := chi.URLParam(request, "size")
		assetInfo.AssocChain = chi.URLParam(request, "chain")
		if assetInfo.AssocChain == "default" {
			assetInfo.AssocChain = ""
		}

		size, err := strconv.ParseFloat(sizeReq, 64)
		if err != nil {
			logrus.Errorf("failed to parse size %s\n", err) // 3.14159265
			render.Render(w, request, transfer.ErrWithdawRender(err))
		}

		assetInfo.Code = currency.NewCode(strings.ToUpper(chi.URLParam(request, "asset")))
		assetInfo.Code.Item.Role = currency.Cryptocurrency

		exchangeEngine, err := d.ExchangeManager.GetExchangeByName(exchangeNameReq)
		if err != nil {
			logrus.Errorf("failed to return exchange %s\n", err)
		}

		wi := &withdraw.Request{
			Exchange: exchangeEngine.GetName(),
			Currency: assetInfo.Code,
			Amount:   size,
			Type:     withdraw.Crypto,
			Crypto: withdraw.CryptoRequest{
				Address:    destinationAddress,
				AddressTag: "",
				FeeAmount:  0,
				Chain:      assetInfo.AssocChain,
			},
		}

		response := transfer.CreateExchangeWithdrawResponse(wi, &d.ExchangeManager)

		logrus.Infof("exchange withdraw response %v", response)
		ctx := context.WithValue(request.Context(), "response", &response)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
