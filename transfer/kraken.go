package transfer

import (
	"github.com/romanornr/autodealer/dealer"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
	"gopkg.in/errgo.v2/fmt/errors"
)

func KrakenConvertUSDTtoEuro() (order.SubmitResponse, error) {
	dealerEngine, err := dealer.NewBuilder().Build()
	if err != nil {
		return order.SubmitResponse{}, err
	}
	exchange, err := dealerEngine.ExchangeManager.GetExchangeByName("Kraken")
	if err != nil {
		return order.SubmitResponse{}, err
	}

	accounts, err := exchange.FetchAccountInfo(asset.Spot)
	if err != nil {
		return order.SubmitResponse{}, errors.Newf("failed to submit order: %s\n", err)
	}

	var value float64

	// check accounts for total tether value to sell
	for _, a := range accounts.Accounts {
		for _, c := range a.Currencies {
			if c.CurrencyName == currency.USDT {
				value = c.TotalValue
			}
		}
	}

	o := &order.Submit{
		Amount:    value,
		Exchange:  exchange.GetName(),
		Type:      order.Market,
		Side:      order.Sell,
		AssetType: asset.Spot,
		Pair:      currency.NewPair(currency.USDT, currency.EUR),
	}

	if value < 10 {
		return order.SubmitResponse{}, errors.Newf("Account doesn't have enough USDT': %f\n", value)
	}

	response, err := exchange.SubmitOrder(o)
	if err != nil {
		return order.SubmitResponse{}, errors.Newf("failed to submit order: %s\n", err)
	}

	logrus.Infof("order response: %v\n", response)
	return response, nil
}

func KrakenInternationalBankAccountWithdrawal() (ExchangeWithdrawResponse, error) {
	dealerEngine, err := dealer.NewBuilder().Build()
	if err != nil {
		return ExchangeWithdrawResponse{}, err
	}
	exchange, err := dealerEngine.ExchangeManager.GetExchangeByName("Kraken")
	if err != nil {
		return ExchangeWithdrawResponse{}, err
	}

	accounts, err := exchange.FetchAccountInfo(asset.Spot)
	if err != nil {
		return ExchangeWithdrawResponse{}, errors.Newf("failed to submit order: %s\n", err)
	}

	var value float64
	for _, a := range accounts.Accounts {
		for _, c := range a.Currencies {
			if c.CurrencyName == currency.EUR {
				value = c.TotalValue
			}
		}
	}

	logrus.Infof("account balance euro before withdraw: %f\n", value)
	if value < 10 {
		return ExchangeWithdrawResponse{}, errors.Newf("The minimal size to withdraw is 10 euro and the current account balance is: %f\n", value)
	}

	baccount, err := dealerEngine.Config.GetExchangeBankAccounts(exchange.GetName(), "romanornr_abn_amro", currency.EUR.String())
	if err != nil {
		logrus.Errorf("failed to get bank account: %v", err)
	}

	var errValid []string
	errValid = baccount.ValidateForWithdrawal(exchange.GetName(), currency.EUR)
	if errValid != nil {
		logrus.Errorf("failed to validate bank account: %v\n", errValid)
	}

	logrus.Infof("baccount %v\n", baccount)

	withdrawRequest := &withdraw.Request{
		Exchange:    exchange.GetName(),
		Currency:    currency.EUR,
		Description: "",
		Amount:      value,
		Type:        withdraw.Fiat,
		Fiat: withdraw.FiatRequest{
			Bank:                          *baccount,
			IsExpressWire:                 true,
			RequiresIntermediaryBank:      false,
			IntermediaryBankAccountNumber: 605,
			IntermediaryBankName:          baccount.BankName,
			IntermediaryBankAddress:       baccount.BankAddress,
			IntermediaryBankCity:          baccount.BankPostalCity,
			IntermediaryBankCountry:       baccount.BankCountry,
			IntermediaryBankPostalCode:    baccount.BankPostalCode,
			IntermediarySwiftCode:         baccount.SWIFTCode,
			IntermediaryBankCode:          baccount.BankCode,
			IntermediaryIBAN:              baccount.IBAN,
			WireCurrency:                  "",
		},
	}

	err = withdrawRequest.Validate()
	if err != nil {
		return ExchangeWithdrawResponse{}, errors.Newf("validation error withdraw request: %s\n", err)
	}

	result := CreateExchangeWithdrawResponse(withdrawRequest, &dealerEngine.ExchangeManager)
	return result, nil
}
