package dealer

import (
	"context"
	"errors"
	"sync"

	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"go.uber.org/multierr"
)

// This code will consist of a single RootStrategy data structure. This data structure would be the only one where Strategy Implementations are held.
// A new RootStrategy object is created. This data structure will be passed down to the last implementation of Strategy.
// After the Strategy Implementations are imported, the order are respective are important.
// The order of import Strategy implementations is important because classes that depend on other classes don't rely on the root strategy in order to work
// so in order to have a non-circular dependency, the classes should be called in a specific order.

// Adding a new strategy is going to be returning two things. The RootStrategy itself, and a specific Strategy that is the type of the Strategy that was requested.
// The ability to add a strategy will create a Map, Map[string] full of RootStrategies. When the end user adds a new strategy, they can give it a unique identifier
// and when they do, we will add a new value to the map. Second, we have a function after the line where a new RootStrategy object is defined; a function called Get
// which takes a string. To get a strategy from a Root Strategy, we simply use this function. We check if the given string matches the String identifier in the Map
// returning the corresponding value in the map. Let's assume similarity with an exchange. Anytime a strategy needs to do something, it will do so through the Dealer,
// which is passing necessary information. Just looking at this graph without talking about the internals, it is easy to see how this system can grow.
// Any additional advanced features or algorithms can be added through additional Function calls, without changing the underlying code of the strategies.

// This dynamic initialization system of functionality has several advantages.
// It allows of the base Dealer to be generic and essentially allow for new additional functionality to be added to it at any time by just adding a new package to the folder structure.
// By using this system it is possible to create a fully fledged Exchange or a versatile universal JSON constructor. Neither of the two situations would be possible using a static approach
// as a static approach leads to a lot of increased code duplication and factoring. This allows for ideas and less explainable of the code.

// Each strategy is instantiated as the variable `m`. Then each of the strategies is passed as a function and called. The function has a Void returned Void.
// Each of the strategy will be called and setup inside their `OnInit()` functions and run their OnInit functions, to set themselves up. Next, each of the OnInit functions will pass a selection of code developers can customize, into their OnFunding() implementations.
// We will then iterate over the variable `rootStrategy` and call each one to execute their function, allowing us to customize each strategy individually.

var (
	ErrStrategyNotFound = errors.New("strategy not found")
	ErrNotStrategy      = errors.New("given object is not a strategy")
)

// RootStrategy is a struct that contains a map of strategies. The map is a sync.Map, which is a thread safe map. The map is initialized with a sync.Map{} and then we can add strategies to it.
// The map is a map of string to Strategy. The string is the name of the strategy and the Strategy is the implementation of the strategy.
type RootStrategy struct {
	strategies sync.Map
}

// NewRootStrategy returns the RootStrategy object. The RootStrategy object has several functions (each).
// creating a RootStrategy variable which is an object. Then we are iterating through each of the additional strategies and adding them to the NewRootStrategy,
// so they are available as the final implementation is created after they have been created
func NewRootStrategy() RootStrategy {
	return RootStrategy{
		strategies: sync.Map{},
	}
}

// Add function takes a string that identifies a implementation of the Strategy, and the implementation of the implementation of the Strategy implementation itself.
// It stores an implementation of a strategy implementation under a string named after the strategy implementation. Which resolves to the correct implementation of the Strategy.
func (m *RootStrategy) Add(name string, s Strategy) {
	m.strategies.Store(name, s)
}

// Delete the Strategy specified by name. You get an object, get the interface's value, and then determine the interface's value.
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

// each function is a function that iterates over all of the current strategies and calls a specific function once for each strategy.
// The closure of the function is the implementation of the Strategy. The function returns an error.
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

// Init function loops through each of the imported Strategy implementations and calls their init functions to initialize them.
// Ordering of implementations is important and if an implementation depends on something another requires you should order the strategy implementations.
func (m *RootStrategy) Init(context context.Context, d *Dealer, e exchange.IBotExchange) error {
	return m.each(func(strategy Strategy) error {
		return strategy.Init(context, d, e)
	})
}

// OnFunding function for the Root strategy. The first line of the function is to call the same function on each_ it is the interface method for the Strategy.
// A new function is called, which is more of an interface for the Strategy called OnFunding. Which allows the user to choose how they want to pass this event to the strategy.
func (m *RootStrategy) OnFunding(d *Dealer, e exchange.IBotExchange, x stream.FundingData) error {
	return m.each(func(strategy Strategy) error {
		return strategy.OnFunding(d, e, x)
	})
}

// The OnPrice implementation of the Strategy is different from above.
// It does not let the user choose how they want to use this information and passes all the information to the specific implementation of that data.
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
