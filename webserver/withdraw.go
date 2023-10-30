package webserver

import (
	"context"
	"github.com/romanornr/autodealer/singleton"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	transfer2 "github.com/romanornr/autodealer/transfer"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
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

// getExchangeWithdrawResponse is the reward for performing an exchange withdraw transaction. It's called
// as part of what is called an exchange event. The received json request is consistent with what is
// expected for what the function defines.
func getExchangeWithdrawResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	exchangeResponse, ok := ctx.Value("response").(*transfer2.ExchangeWithdrawResponse) // TODO fix
	if !ok {
		logrus.Errorf("Got unexpected response %T instead of *ExchangeWithdrawResponse", exchangeResponse)
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		render.JSON(w, r, http.StatusUnprocessableEntity)
		return
	}
	// render.Render(w, r, exchangeResponse)
	render.JSON(w, r, exchangeResponse)
}

// WithdrawCtx is an HTTP handler function which stores the request input with the help of chi.URLParams get method
// in the response and call the createExchangeWithdrawResponse to create an exchange withdrawal transaction
// for the specified exchange
func WithdrawCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var err error
		assetInfo := new(Asset)

		d, err := singleton.GetDealer(context.Background()) //d, err := dealer.NewBuilder().Build()
		if err != nil {
			logrus.Errorf("failed to create a dealer %s\n", err)
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		}

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
			logrus.Errorf("failed to convert size %s\n", err)
			render.Status(request, http.StatusUnprocessableEntity)
			render.JSON(w, request, http.StatusUnprocessableEntity)
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

		response, err := transfer2.CreateExchangeWithdrawResponse(wi, exchangeEngine)
		if err != nil {
			render.JSON(w, request, ErrRender(err))
		}

		logrus.Infof("exchange withdraw response %v\n", response)
		ctx := context.WithValue(request.Context(), "response", &response)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
