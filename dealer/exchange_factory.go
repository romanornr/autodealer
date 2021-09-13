package dealer

import (
	"errors"

	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
)

var ErrCreatorNotRegistered = errors.New("creator not registered")

type (
	ExchangeCreatorFunc func() (exchange.IBotExchange, error)

	// ExchangeFactory map[string]ExchangeCreatorFunc
	// A higher order function, a "Factory method" over a map, where you cann register a function to a particular key of a map.
	ExchangeFactory map[string]ExchangeCreatorFunc
)

// exchange_factory["BTC-e"] = BTCe.Create
// Then whenever you do, exchange_factory["ftx"]()
// You get an instantiated `ftx` bot exchange. Easy way to make all sorts of mini plugin like things, i.e. bots, but also
// decouple them from main code and make it very simple to add new ones.
// - `name`: the name of the exchange.
// - `what`: the name of the function to register.
// - `factory`: an ExchangeFactory, generally massaged using the Forge(...) call.

// Register is a mechanism to allow an exchange to register with a visibility broker e is the broker's exchange factory.
// name is the name of the exchange (should be unique)
// check is a channel for returning results
// rule: rule returning the channel to channel for this exchange channel
func (e ExchangeFactory) Register(name string, fn ExchangeCreatorFunc) {
	e[name] = fn
}

// NewExchangeByName creates an exchange based on the given string.
func (e ExchangeFactory) NewExchangeByName(name string) (exchange.IBotExchange, error) {
	fn, ok := e[name]
	if !ok {
		return nil, ErrCreatorNotRegistered
	}

	// return fn()

	// Get the newly created exchange
	newExchange, err := fn()
	if err != nil {
		return nil, err
	}

	return newExchange, err
}
