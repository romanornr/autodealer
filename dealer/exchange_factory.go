package dealer

import (
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
)

// ExchangeFactory defines an interface for creating exchange instances.
// It abstracts the instantiation process, allowing for flexible and dynamic creation of different exchanges based on a given name.
type ExchangeFactory interface {
	NewExchangeByName(name string) (exchange.IBotExchange, error)
}
