package dealer

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"strings"
	"sync"
	"time"
)

var (
	ErrAccountIndexOfRange = errors.New("no account with this index exist")
	ErrCurrencyNotFound    = errors.New("currency not found in holdings")
	ErrHoldingsNotFound    = errors.New("holdings not found")
)

//The code is very simple; it's mostly straightforward concurrency control (use sync.Mutex for resource access),
//error handling, and logging. A strategy is composed of two components: TickerStrategy and the BalancesStrategy itself.

//First, the TickerStrategy is initialised to wait for `refreshRate` minutes before triggering an event on its own.
//Define the tick function defined on the TickerStrategy to do the refreshing of the account balances after `refreshRate`
//minutes, and hook this to its own ticker's `Tick()` method. All ExchangeExchange could do is initiate an event for an exchange.
//All other logic around fetching balance info prior to ticker had to be handled by BalancesUpdateTicker.

//The ticker triggered event will now be handled by the BalancesStrategy. First, create an empty ExchangeHoldings instance
//and load the individual account balances for each asset type the particular exchange supports. Populating the
//ExchangeHoldings instance can be done concurrently by fetching info for each asset type concurrently. Reject any empty
//results.

// BalancesStrategy struct maps exchange names as keys to their respective exchange holdings as values.
// Using this strategy, we subscribe to a user defined number of tickers and use those to update our holdings.
// We then atomically update the exchange holding map in our balance strategy struct, and we can read an `ExchangeHoldings` object at any point.
// Note that  this strategy does not order, so stops and take profits can't be engaged. To use it, initialize an exchange and a builder and then simply
type BalancesStrategy struct {
	holdings sync.Map
	ticker   TickerStrategy
}

// NewBalancesStrategy function creates an instance of the BalancesStrategy struct. In turn, the BalancesStrategy struct creates a TickFunc method as a `ticker.Ticker.TickFunc`
// Then assigns that TickFunc function to the TickerStrategy object that is to be used later.
func NewBalancesStrategy(refreshRate time.Duration) Strategy {
	b := &BalancesStrategy{
		holdings: sync.Map{},
		ticker: TickerStrategy{
			Interval: refreshRate,
			TickFunc: nil,
			tickers:  sync.Map{},
		},
	}
	b.ticker.TickFunc = b.tick
	return b
}

// ExchangeHoldings method returns all the balance from an exchange including all the global account information and information for each asset and account.
func (b *BalancesStrategy) ExchangeHoldings(exchangeName string) (*ExchangeHoldings, error) {
	key := strings.ToLower(exchangeName)

	if ptr, ok := b.holdings.Load(key); ok {
		if h, ok := ptr.(*ExchangeHoldings); ok {
			return h, nil
		}
	}
	return nil, ErrHoldingsNotFound
}

// tick method (executed via the tickFunc member in dealer.TickerStrategy) for a given exchange's ticker takes the interval set in the strategy options
// which by default should be a constant time of x seconds, and retrieves all holding information for all tickers, currency and asset types on the exchange.
func (b *BalancesStrategy) tick(d *Dealer, e exchange.IBotExchange) {
	// create a new holdings' struct that we'll fill out and then atomically update
	holdings := NewExchangeHoldings()

	// go through all the asset types, fetch account info for each of them and aggregate them into dealer.Holdings
	for _, assetType := range e.GetAssetTypes(true) {
		h, err := e.UpdateAccountInfo(context.Background(), assetType)
		if err != nil {
			logrus.Errorf("exchange %s: %s\n", e.GetName(), err)
			continue
		}

		for _, subAccount := range h.Accounts {
			if _, ok := holdings.Accounts[subAccount.ID]; !ok {
				holdings.Accounts[subAccount.ID] = SubAccount{
					ID:       subAccount.ID,
					Balances: make(map[asset.Item]map[currency.Code]CurrencyBalance),
				}
			}
			if _, ok := holdings.Accounts[subAccount.ID].Balances[assetType]; ok {
				holdings.Accounts[subAccount.ID].Balances[assetType] = make(map[currency.Code]CurrencyBalance)
			}
			for _, currencyBalance := range subAccount.Currencies {
				holdings.Accounts[subAccount.ID].Balances[assetType][currencyBalance.CurrencyName] = CurrencyBalance{
					Currency:   currencyBalance.CurrencyName,
					TotalValue: currencyBalance.TotalValue,
					Hold:       currencyBalance.Hold,
				}
			}
		}
		key := strings.ToLower(e.GetName())
		b.holdings.Store(key, holdings)
	}
}

// +--------------------+
// | Strategy interface |
// +--------------------+
// the Strategy interface keeps track of the portfolios of exchange the bot is connected to

// Init method loads an initial holdings item and then calls the internal tickers Init.
// This merges well with the LazyHandler holding the last refresh timestamp since the refresh should already be running when this is called.
func (b *BalancesStrategy) Init(ctx context.Context, d *Dealer, e exchange.IBotExchange) error {
	key := strings.ToLower(e.GetName())
	b.holdings.Store(key, NewExchangeHoldings())

	return b.ticker.Init(ctx, d, e)
}

// OnFunding method is not being used because the backtest loads directly from a saved ledger
func (b *BalancesStrategy) OnFunding(d *Dealer, e exchange.IBotExchange, x stream.FundingData) error {
	return nil
}

// OnPrice method is called every time a new price is received from the exchange.
func (b *BalancesStrategy) OnPrice(d *Dealer, e exchange.IBotExchange, x ticker.Price) error {
	return nil
}

// OnKline is called when a new kline is received.
func (b *BalancesStrategy) OnKline(d *Dealer, e exchange.IBotExchange, x stream.KlineData) error {
	return nil
}

func (b *BalancesStrategy) OnOrderBook(d *Dealer, e exchange.IBotExchange, x orderbook.Base) error {
	return nil
}

func (b *BalancesStrategy) OnOrder(d *Dealer, e exchange.IBotExchange, x order.Detail) error {
	return nil
}

func (b *BalancesStrategy) OnModify(d *Dealer, e exchange.IBotExchange, x order.Modify) error {
	return nil
}

func (b *BalancesStrategy) OnBalanceChange(d *Dealer, e exchange.IBotExchange, x account.Change) error {
	return nil
}

func (b *BalancesStrategy) OnUnrecognized(d *Dealer, e exchange.IBotExchange, x interface{}) error {
	return nil
}

// Deinit should be called when the bot is shutting down or closing functions given by the exchange's API interface.It is used to stop the ticker.
func (b *BalancesStrategy) Deinit(d *Dealer, e exchange.IBotExchange) error {
	return b.ticker.Deinit(d, e)
}
