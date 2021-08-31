package webserver

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/romanornr/autodealer/dealer"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
	"gopkg.in/errgo.v2/fmt/errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// exchangewithdrawResponse is a struct that is designed to represent the response from the ExchangesWithdrawals API call. ExchangesWithdrawals is a simple function which returns deposit, withdraw, trade and withdrawal information so we will only add the information there which we are interested in:
// ExchangesResponse is a struct that includes information about the request as well as the response, we're only interested in the response hence why we've added resp.
// We've used the Withdrawal struct as that is the response from the exchange (withdraw.ExchangeResponse). The Exchange key is the exchange used to make the request.
// The Type key represents the type of information requested in the Call function. The DestinationAddress is the address the withdrawal was sent to if the request used the DepositAddress field.
// The Time key is when the request was made and the Err field returns errors if any occurred.
type exchangeWithdrawResponse struct {
	ExchangeResponse   *withdraw.ExchangeResponse
	Exchange           string               `json:"exchange"`
	Type               withdraw.RequestType `json:"type"`
	DestinationAddress string               `json:"destination"`
	Time               time.Time            `json:"time"`
	Err                error                `json:"err"`
}

// createExchangeWithdrawResponse function creates a withraw request using exchangeManager and returns the exchangeWithdrawResponse including response
// It first creates an exchange manager by name which will fetch the exchange name from the engine.
//This function will fetch the exchange details from the exchange name and returns the balance of an asset for a user.
// Next it creates the WithdrawManager for a given exchange, tries to withdraw the crypto asset, and gets the destination address. This is done by calling the withdraw crypto trade function
// so here's the thing  this function returns an Exchange response which holds the deposit id  on that exchange.
// Finally, we update the results which we return in JSON format.
// After we make sure that the withdrawal functionality is working we can inject the functionality in the withdrawal method of the engine struct.
func createExchangeWithdrawResponse(withdrawRequest *withdraw.Request, exchangeManager *engine.ExchangeManager) exchangeWithdrawResponse { // withdrawManager *engine.WithdrawManager) exchangeWithdrawResponse {
	manager, err := exchangeManager.GetExchangeByName(withdrawRequest.Exchange)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to create exchangeManager by name %s\n", err))
	}

	var exchangeResponse *withdraw.ExchangeResponse

	var response = exchangeWithdrawResponse{
		ExchangeResponse: exchangeResponse,
		Exchange:         manager.GetName(),
		Type:             withdrawRequest.Type,
		Err:              err,
	}

	if withdrawRequest.Type == withdraw.Crypto {
		response.ExchangeResponse, err = manager.WithdrawCryptocurrencyFunds(withdrawRequest)
		if err != nil {
			err = errors.New(fmt.Sprintf("failed to withdraw crypto asset %s\n", err))
		}
		response.DestinationAddress = withdrawRequest.Crypto.Address
	}
	return response
}

// Render exchangeWithdrawResponse implements the error interface to show the user an error occured if exchangeWithdrawRequest returns an error.
func (e exchangeWithdrawResponse) Render(w http.ResponseWriter, r *http.Request) error {
	e.Time = time.Now()
	return e.Err
}

// ErrWithdawRender as JSON if err is not nil.
// If err is nil, then Render http.StatusOK. If err then Render an Error response if it implements AbsError we log the error message.
// If it does not implement AbsError we log to err type.
func ErrWithdawRender(err error) render.Renderer {
	return &exchangeWithdrawResponse{
		Err: err,
	}
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
	exchangeResponse, ok := ctx.Value("response").(*exchangeWithdrawResponse) // TODO fix
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		render.Render(w, r, ErrWithdawRender(errors.Newf("Failed to renderWithdrawResponse")))
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

		dealer, err := dealer.New(engine.Settings{})
		if err != nil {
			logrus.Errorf("failed to create dealer %s\n", err)
		}

		exchangeNameReq := chi.URLParam(request, "exchange")
		destinationAddress := chi.URLParam(request, "destinationAddress")
		sizeReq := chi.URLParam(request, "size")

		size, err := strconv.ParseFloat(sizeReq, 64)
		if err != nil {
			logrus.Errorf("failed to parse size %s\n", err) // 3.14159265
			render.Render(w, request, ErrWithdawRender(err))
		}

		assetInfo.Code = currency.NewCode(strings.ToUpper(chi.URLParam(request, "asset")))
		assetInfo.Code.Item.Role = currency.Cryptocurrency

		exchangeEngine, err := dealer.ExchangeManager.GetExchangeByName(exchangeNameReq)
		if err != nil {
			logrus.Errorf("Failed to return exchange %s\n", err)
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
			},
		}

		response := createExchangeWithdrawResponse(wi, &dealer.ExchangeManager)

		logrus.Infof("exchange withdraw response %v", response)
		ctx := context.WithValue(request.Context(), "response", &response)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
