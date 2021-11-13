package dealer

import (
	"errors"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
)

var ErrCreatorNotRegistered = errors.New("exchange creator not registered")

type (
	ExchangeFactory func(name string) (exchange.IBotExchange, error)
)

// NewExchangeByName implements gocryptotrader/engine.CustomExchangeBuilder.
func (e ExchangeFactory) NewExchangeByName(name string) (exchange.IBotExchange, error) {
	return e(name)
}
