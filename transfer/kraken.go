package transfer

import (
	"context"
	"github.com/romanornr/autodealer/dealer"

	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
	"gopkg.in/errgo.v2/fmt/errors"
)

// KrakenConvertUSDT converts all USDT to Euros
func KrakenConvertUSDT(code currency.Code, d *dealer.Dealer) (order.SubmitResponse, error) {

	exchange, err := d.ExchangeManager.GetExchangeByName("Kraken")
	if err != nil {
		return order.SubmitResponse{}, err
	}

	accounts, err := exchange.FetchAccountInfo(context.Background(), asset.Spot)
	if err != nil {
		return order.SubmitResponse{}, errors.Newf("failed to submit order: %s\n", err)
	}

	var value float64

	// check accounts for total tether value to sell
	for _, a := range accounts.Accounts {
		for _, c := range a.Currencies {
			if c.CurrencyName == currency.USDT {
				value = c.Total
			}
		}
	}

	//currency.NewPair(currency.USDT, code)

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

	response, err := exchange.SubmitOrder(context.Background(), o)
	if err != nil {
		return order.SubmitResponse{}, errors.Newf("failed to submit order: %s\n", err)
	}

	logrus.Infof("order response: %v\n", response)
	return *response, nil
}

// KrakenInternationalBankAccountWithdrawal withdraws funds to an international bank account
func KrakenInternationalBankAccountWithdrawal(code currency.Code, d *dealer.Dealer) (ExchangeWithdrawResponse, error) {

	exchange, err := d.ExchangeManager.GetExchangeByName("Kraken")
	if err != nil {
		return ExchangeWithdrawResponse{}, err
	}

	accounts, err := exchange.FetchAccountInfo(context.Background(), asset.Spot)
	if err != nil {
		return ExchangeWithdrawResponse{}, errors.Newf("failed to submit order: %s\n", err)
	}

	var value float64
	for _, a := range accounts.Accounts {
		for _, c := range a.Currencies {
			if c.CurrencyName == currency.EUR {
				value = c.Total
			}
		}
	}

	logrus.Infof("account balance euro before withdraw: %f\n", value)
	if value < 10 {
		err = errors.Newf("The minimal size to withdraw is 10 euro and the current account balance is: %f\n", value)
		return ExchangeWithdrawResponse{Error: err}, err
	}

	baccount, err := d.Config.GetExchangeBankAccounts(exchange.GetName(), "romanornr_abn_amro", code.String())
	if err != nil {
		logrus.Errorf("failed to get bank account: %v", err)
	}

	var errValid []string
	errValid = baccount.ValidateForWithdrawal(exchange.GetName(), code)
	if errValid != nil {
		logrus.Errorf("failed to validate bank account: %v\n", errValid)
	}

	logrus.Infof("baccount %v\n", baccount)

	withdrawRequest := &withdraw.Request{
		Exchange:    exchange.GetName(),
		Currency:    code,
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

	if err = withdrawRequest.Validate(); err != nil {
		return ExchangeWithdrawResponse{}, errors.Newf("validation error withdraw request: %s\n", err)
	}

	result, err := CreateExchangeWithdrawResponse(withdrawRequest, exchange)
	if err != nil {
		logrus.Errorf("failed to create withdraw response: %v\n", err)
	}
	return result, err
}
