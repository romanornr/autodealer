package subaccount

import (
	"context"
	"github.com/sirupsen/logrus"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

// GetByID is a function that returns a subaccount by ID.
func GetByID(e exchange.IBotExchange, accountId string) (account.SubAccount, error) {
	accounts, err := e.UpdateAccountInfo(context.Background(), asset.Spot)
	if err != nil {
		logrus.Errorf("failed to get exchange account: %s\n", err)
	}

	// return the first account if there's no other accounts
	if len(accounts.Accounts) == 1 {
		return accounts.Accounts[0], nil
	}

	for _, a := range accounts.Accounts {
		// return the main account for FTX
		if a.ID == "main" && e.GetName() == "FTX" {
			return a, nil
		}

		if a.ID == accountId {
			return a, nil
		}
	}
	return account.SubAccount{}, err
}
