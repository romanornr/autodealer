package dealer

import (
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

// CurrencyBalance struct is an easy way to house the pairs of currency held.
// struct with our most used currencies. Just with some nice printing methods for the most part.
type CurrencyBalance struct {
	Currency currency.Code
	TotalValue float64
	Hold float64
}

// SubAccount struct is an easy way to group our connected accounts. To hold all connected exchanges.
type SubAccount struct {
	ID string
	Balances map[asset.Item]map[currency.Code]CurrencyBalance
}

// ExchangeHoldings struct is a struct where we house a map[string]SubAccount.
// For accounts, we basically house all our accounts now as a string as they are all now connected to gct now as a linked service. To make them easier to reference.
type ExchangeHoldings struct {
	Accounts map[string]SubAccount
}

// NewExchangeHoldings function is an easy way to create an empty ExchangeHoldings struct, so we can create an empty struct on startup to avoid us facing gct/goat by ensuring state on startup
func NewExchangeHoldings() *ExchangeHoldings {
	return &ExchangeHoldings{
		Accounts: make(map[string]SubAccount),
	}
}

// CurrencyBalance method is just a simple way to conform and enforce that we pass
// a exchange and an item, and we use them to match what we store and retrieve.
func (h *ExchangeHoldings) CurrencyBalance(accountID string, code currency.Code, asset asset.Item)  (CurrencyBalance, error) {
	account, ok := h.Accounts[accountID]
	if !ok {
		var empty CurrencyBalance
		return empty, ErrCurrencyNotFound
	}
	c, ok := account.Balances[asset][code]
	if !ok {
		var empty CurrencyBalance
		return empty, ErrCurrencyNotFound
	}
	return c, nil
}