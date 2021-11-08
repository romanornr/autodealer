package transfer

import (
	"context"
	"fmt"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kraken"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
	"net/http"
	"time"
)

// ExchangeWithdrawResponse is a struct that is designed to represent the response from the ExchangesWithdrawals API call. ExchangesWithdrawals is a simple function which returns deposit, withdraw, trade and withdrawal information so we will only add the information there which we are interested in:
// ExchangesResponse is a struct that includes information about the request as well as the response, we're only interested in the response hence why we've added resp.
// We've used the Withdrawal struct as that is the response from the exchange (withdraw.ExchangeResponse). The Exchange key is the exchange used to make the request.
// The Type key represents the type of information requested in the Call function. The DestinationAddress is the address the withdrawal was sent to if the request used the DepositAddress field.
// The Time key is when the request was made and the Err field returns errors if any occurred.
type ExchangeWithdrawResponse struct {
	ExchangeResponse   *withdraw.ExchangeResponse
	Exchange           string               `json:"exchange"`
	Type               withdraw.RequestType `json:"type"`
	DestinationAddress string               `json:"destination"`
	Time               time.Time            `json:"time"`
	Error              error                `json:"error"`
	Success            bool                 `json:"success"`
}

// Render implement rest.Render to render the Response of the ExchangeWithdraw call
func (e ExchangeWithdrawResponse) Render(w http.ResponseWriter, r *http.Request) error {
	e.Time = time.Now()
	return e.Error
}

// ErrWithdawRender as JSON if err is not nil.
// If err is nil, then Render http.StatusOK. If err then Render an Error response if it implements AbsError we log the error message.
// If it does not implement AbsError we log to err type.
func ErrWithdawRender(err error) render.Renderer {
	return &ExchangeWithdrawResponse{
		Error: err,
	}
}

// CreateExchangeWithdrawResponse function creates a withraw request using exchangeManager and returns the exchangeWithdrawResponse including response
// It first creates an exchange manager by name which will fetch the exchange name from the engine.
// This function will fetch the exchange details from the exchange name and returns the balance of an asset for a user.
// Next it creates the WithdrawManager for a given exchange, tries to withdraw the crypto asset, and gets the destination address. This is done by calling the withdraw crypto trade function
// so here's the thing  this function returns an Exchange response which holds the deposit id  on that exchange.
// Finally, we update the results which we return in JSON format.
// After we make sure that the withdrawal functionality is working we can inject the functionality in the withdrawal method of the engine struct.
func CreateExchangeWithdrawResponse(withdrawRequest *withdraw.Request, exchangeManager exchange.IBotExchange) (ExchangeWithdrawResponse, error) { // withdrawManager *engine.WithdrawManager) exchangeWithdrawResponse {
	var exchangeResponse *withdraw.ExchangeResponse
	var err error
	logrus.Info("creating withdraw response for exchange")

	if withdrawRequest.Type == withdraw.Crypto {
		exchangeResponse, err = exchangeManager.WithdrawCryptocurrencyFunds(context.Background(), withdrawRequest)
		if err != nil {
			err = fmt.Errorf("failed to withdraw crypto asset %s\n", err)
		}
		logrus.Infof("exchange response: %v\n", exchangeResponse)
	}

	if withdrawRequest.Type == withdraw.Fiat && exchangeManager.GetName() == "Kraken" {
		logrus.Info("withdraw kraken")
		k := kraken.Kraken{Base: *exchangeManager.GetBase()}
		exchangeResponse.ID, err = k.Withdraw(context.Background(), currency.EUR.String(), withdrawRequest.Fiat.Bank.ID, withdrawRequest.Amount)
		if err != nil {
			err = fmt.Errorf("failed international bank withdraw request: %s\n", err)
		}
	}

	var response = ExchangeWithdrawResponse{
		ExchangeResponse: exchangeResponse,
		Exchange:         exchangeManager.GetName(),
		Type:             withdrawRequest.Type,
		//DestinationAddress: withdrawRequest.Crypto.Address,
		Time:  time.Now(),
		Error: err,
		Success: false,
	}

	if withdrawRequest.Type == withdraw.Crypto {
		response.DestinationAddress = withdrawRequest.Crypto.Address
	}

	if err == nil {
		response.Success = true
	}

	return response, err
}
