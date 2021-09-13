package dealer

import (
	"errors"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"go.uber.org/multierr"
	"sync"
)

// Adding a new strategy is going to be returning two things. The RootStrategy itself, and a specific Strategy that is the type of the Strategy that was requested.
// The ability to add a strategy will create a Map, Map[string] full of RootStrategies. When the end user adds a new strategy, they can give it a unique identifier
// and when they do, we will add a new value to the map. Second, we have a function after the line where a new RootStrategy object is defined; a function called Get
// which takes a string. To get a strategy from a Root Strategy, we simply use this function. We check if the given string matches the String identifier in the Map
// returning the corresponding value in the map. Let's assume similarity with an exchange. Anytime a strategy needs to do something, it will do so through the Dealer,
// which is passing necessary information. Just looking at this graph without talking about the internals, it is easy to see how this system can grow.
// Any additional advanced features or algorithms can be added through additional Function calls, without changing the underlying code of the strategies.

var (
	ErrStrategyNotFound = errors.New("strategy not found")
	ErrNotStrategy      = errors.New("given object is not a strategy")
)

//RootStrategy is a strategy implementation
type RootStrategy struct {
	strategies sync.Map
}

// NewRootStrategy is a constructor for a Stock Exchange
func NewRootStrategy() RootStrategy {
	return RootStrategy{
		strategies: sync.Map{},
	}
}

// Add inserts a strategy with a specific name
func (m *RootStrategy) Add(name string, s Strategy) {
	m.strategies.Store(name, s)
}

// Delete removes a strategy with a
func (m *RootStrategy) Delete(name string) (Strategy, error) {
	x, ok := m.strategies.LoadAndDelete(name)
	if !ok {
		return nil, ErrStrategyNotFound
	}
	return x.(Strategy), nil
}

// Get returns the strategy with the given name
func (m *RootStrategy) Get(name string) (Strategy, error) {
	x, ok := m.strategies.Load(name)
	if !ok {
		return nil, ErrStrategyNotFound
	}
	return x.(Strategy), nil
}

// each iterates over each Strategy, calling Function f once per Strategy
// Returns nil on success, or Function specific error on failure
func (m *RootStrategy) each(f func(Strategy) error) error {
	var err error

	m.strategies.Range(func(key, value interface{}) bool {
		s, ok := value.(Strategy)
		if !ok {
			err = multierr.Append(err, ErrNotStrategy)
		} else {
			err = multierr.Append(err, f(s))
		}
		return true
	})
	return err
}

// Init Initialize strategies of Dealer
func (m *RootStrategy) Init(d *Dealer, e exchange.IBotExchange) error {
	return m.each(func(strategy Strategy) error {
		return strategy.Init(d, e)
	})
}

// OnFunding Each strategy is called for each Funding order. When all your
// strategies have been applied, it returns a list of errors from the
// Apply methods of the returned strategies
func (m *RootStrategy) OnFunding(d *Dealer, e exchange.IBotExchange, x stream.FundingData) error {
	return m.each(func(strategy Strategy) error {
		return strategy.OnFunding(d, e, x)
	})
}

// OnPrice is called whenever a price update is published for a ticker
func (m *RootStrategy) OnPrice(d *Dealer, e exchange.IBotExchange, x ticker.Price) error {
	return m.each(func(strategy Strategy) error {
		return strategy.OnPrice(d, e, x)
	})
}

// OnKline listens to the Kline stream data events and execute optional action
func (m *RootStrategy) OnKline(d *Dealer, e exchange.IBotExchange, x stream.KlineData) error {
	return m.each(func(strategy Strategy) error {
		return strategy.OnKline(d, e, x)
	})
}

// OnOrderBook is called when initial orderbook is created or if the bids or asks change
// Must pass in Dealer that created this strategy. Also pass in the Exchange used by the strategy
// Call this function once per Strategy
func (m *RootStrategy) OnOrderBook(d *Dealer, e exchange.IBotExchange, x orderbook.Base) error {
	return m.each(func(strategy Strategy) error {
		return strategy.OnOrderBook(d, e, x)
	})
}

// OnOrder is called when changes occur to a specific order
func (m *RootStrategy) OnOrder(d *Dealer, e exchange.IBotExchange, x order.Detail) error {
	return m.each(func(strategy Strategy) error {
		return strategy.OnOrder(d, e, x)
	})
}

// OnModify is invoked when an order is modified.
// The arguments passed are the original user message
func (m *RootStrategy) OnModify(d *Dealer, e exchange.IBotExchange, x order.Modify) error {
	return m.each(func(strategy Strategy) error {
		return strategy.OnModify(d, e, x)
	})
}

// OnBalanceChange iterates over each strategy, calling OnBalanceChange, logging an error if any fail
// Returns nil on success, or Function specific error on failure
func (m *RootStrategy) OnBalanceChange(d *Dealer, e exchange.IBotExchange, x account.Change) error {
	return m.each(func(strategy Strategy) error {
		return strategy.OnBalanceChange(d, e, x)
	})
}

// OnUnrecognized is called on unrecognized data
func (m *RootStrategy) OnUnrecognized(d *Dealer, e exchange.IBotExchange, x interface{}) error {
	return m.each(func(strategy Strategy) error {
		return strategy.OnUnrecognized(d, e, x)
	})
}

// Deinit deinitializes strategies in a specific Dealer struct
// For each strategy in a Dealer, calls Strategy.Deinit()
func (m *RootStrategy) Deinit(d *Dealer, e exchange.IBotExchange) error {
	return m.each(func(strategy Strategy) error {
		return strategy.Deinit(d, e)
	})
}
