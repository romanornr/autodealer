package transfer

import (
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kraken"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/portfolio/banking"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
	"gopkg.in/errgo.v2/fmt/errors"
)

func KrakenConvertUSDTtoEuro() (order.SubmitResponse, error) {
	krakenEngine := engine.Bot.GetExchangeByName("Kraken")

	accounts, err := krakenEngine.FetchAccountInfo(asset.Spot)
	if err != nil {
		return order.SubmitResponse{}, errors.Newf("failed to submit order: %s\n", err)
	}

	account := accounts.Accounts[0]
	var value float64

	for _, c := range account.Currencies {
		if c.CurrencyName == currency.USDT {
			value = c.TotalValue
		}
	}

	o := &order.Submit{
		Amount:    value,
		Exchange:  krakenEngine.GetName(),
		Type:      order.Market,
		Side:      order.Sell,
		AssetType: asset.Spot,
		Pair:      currency.NewPair(currency.USDT, currency.EUR),
	}

	if value < 10 {
		return order.SubmitResponse{}, errors.Newf("Account doesn't have enough USDT': %f\n", value)
	}

	response, err := krakenEngine.SubmitOrder(o)
	if err != nil {
		return order.SubmitResponse{}, errors.Newf("failed to submit order: %s\n", err)
	}

	logrus.Infof("order response: %v\n", response)
	return response, nil
}

func KrakenInternationalBankAccountWithdrawal() (string, error) {
	krakenEngine := engine.Bot.GetExchangeByName("Kraken")
	accounts, err := krakenEngine.FetchAccountInfo(asset.Spot)
	if err != nil {
		return "", errors.Newf("failed to submit order: %s\n", err)
	}

	account := accounts.Accounts[0]
	var value float64

	for _, c := range account.Currencies {
		if c.CurrencyName == currency.EUR {
			value = c.TotalValue
		}
	}

	logrus.Infof("account balance euro before withdraw: %f\n", value)
	if value < 10 {
		return "", errors.Newf("The minimal size to withdraw is 10 euro and the current account balance is: %f\n", value)
	}


	//bankAccounts := engine.Bot.Config.BankAccounts[0].ID
	baccount, err  := banking.GetBankAccountByID("romanornr_abn_amro")
	if err != nil {
		logrus.Errorf("failed to get bank account: %v", err)
	}
	logrus.Info(baccount.ValidateForWithdrawal("kraken", currency.EUR))

	logrus.Infof("baccount %v\n", baccount)

	withdrawRequest := &withdraw.Request{
		Exchange:    krakenEngine.GetName(),
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
		return "", errors.Newf("validation error withdraw request: %s\n", err)
	}

	k := kraken.Kraken{ Base: *krakenEngine.GetBase()}
	result, err := k.Withdraw(currency.EUR.String(), baccount.ID, value)
	if err != nil {
		return "", errors.Newf("failed international bank withdraw request: %s\n", err)
	}

	return result, nil
}
