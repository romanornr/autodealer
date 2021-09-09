package dealer

import (
	"errors"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"go.uber.org/multierr"
	"sync"
)

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
			err = multierr.Append(err, ErrStrategyNotFound)
		} else {
			err = multierr.Append(err, f(s))
		}
		return true
	})
	return err
}

// Init Initialize strategies of Dealer
func (m *RootStrategy) Init(d *Dealer, e exchange.IBotExchange) error {
	return m.each(func(s Strategy) error { return s.Init(d, e) })
}
