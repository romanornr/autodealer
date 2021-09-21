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

// This code is helpful in development and production environments when a large number of exchanges are created.
// Then it allows you to create a single instance of an exchange by specifying the exchange type.
// If you had another bot on another platform, you could create an instance of the right type of exchange, but with out requiring the new platform to be written yet

// This approach is that we can create new exchanges very easily and quick. All we need to do is register a new exchange, and we can immediately use it.
// That means we can quickly pass around and reference exchanges and quickly and easily use them in our system
// And we don't need to change any other code if we want to add or remove new exchanges because this factory allows us to register any exchange.

// Register function creates a key-value pair in the ExchangeFactory that is in the form of `key` = name of exchange`value` = A reference to the function named fn.
// In this instance, the convention is name, the exchange type in uppercase, followed by Exchange lowercase.
func (e ExchangeFactory) Register(name string, fn ExchangeCreatorFunc) {
	e[name] = fn
}

// NewExchangeByName does a soft interpretation by determining if the exchange name requested matches one of the factory's stated functions and then initiating a connection to it.
// If not, we instantly search for the name in the worldwide list of exchanges.
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
