package dealer

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"sync"
	"time"
)

var (
	ErrAccountIndexOfRange = errors.New("no account with this index exist")
	ErrCurrencyNotFound    = errors.New("currency not found in holdings")
	ErrHoldingsNotFound    = errors.New("holdings not found")
)

// purpose is converting between the internal ticker type `TickerStrategy` with it's associated `Dealer` and exchange.IBotExchange
// and the generic `Strategy` interface with it's associated `Dealer` and pretty much any interface.

// The code implements a simple load balancing technique and binds all transactions to a single exchange for an associated ticker ( Interval , ticker type etc ).
// Underneath, a ticker tickerStrategy struct keeps a reference to the supplied dealer, keeps a reference of the exchange for dealing requests associated
// Additionally, it maintains a map of tickerable intervals to their associated tickers for this range of desired frequencies.
// We use context.Background() to close out this request by registering a timeout. When things are done under the tickerStrategy functionality.

// BalancesStrategy primarily intended to facilitate the process of calculating the value of our coin equivalents in an atomic and efficient manner.
// The Load and Store operations let us get and set ( or, if desired, concurrently retrieve and set ) our Holdings.
// The primary activity that we are concerned with is cross-checking all of our current accounts to ensure that they are in accordance with our intended holdings.
// BalancesStrategy struct initialises to nil, keeps reference of the associated `TickerStrategy` struct and ensures the `TickerStrategy` receives an initial value.
// The ticker struct contains information related to its own `Interval`, `TickFunc`, associated `dealer` and `exchanges`.
type BalancesStrategy struct {
	balances sync.Map
	ticker   TickerStrategy
}

// NewBalancesStrategy creates a new instance of the `BalancesStrategy` struct.
// It takes in a `refreshRate` parameter which is the time interval in which the balances are refreshed.
func NewBalancesStrategy(refreshRate time.Duration) Strategy {
	b := &BalancesStrategy{
		balances: sync.Map{},
		ticker: TickerStrategy{
			Interval: refreshRate,
			TickFunc: nil,
			tickers:  sync.Map{},
		},
	}
	b.ticker.TickFunc = b.tick
	return b
}

// tick creates a basic form of load balancing. If the ticker type strategy has already been created for this exchange
// then no action will be taken because the orders are still submit through the existing ticker type strategy.
// All the TickFunc check ensures that all balances happen more or less at the same time.
// A periodic check of the accounts' info avoids the chances of a new ticker taking a huge load off of a single one of the secrets.
func (b *BalancesStrategy) tick(d *Dealer, e exchange.IBotExchange) {
	holdings, err := e.UpdateAccountInfo(context.Background(), asset.Spot)
	if err != nil {
		logrus.Errorf("exchange. %s\n", e.GetName())
	} else {
		b.Store(holdings)
	}
}

// Store stores holdings information into the balances strategy
func (b *BalancesStrategy) Store(holdings account.Holdings) {
	b.balances.Store(holdings.Exchange, holdings)
}

// Load returns an amortized holding from self.holdings of exchange exchangeName of the accountID of the key code of the asset you want
// in exchange to go by. In loading balance of account accountID of base asset name of code you risk a panic if there is a lack
// of a BaseBalance on exchange by accountID for base asset name of code.
func (b *BalancesStrategy) Load(exchangeName string) (holdings account.Holdings, loaded bool) {
	var ok bool
	pointer, loaded := b.balances.Load(exchangeName)
	if loaded {
		holdings, ok = pointer.(account.Holdings)
		if !ok {
			logrus.Panicf("have %T, want account.Holdings", pointer)
		}
	}
	return holdings, ok
}

// Currency returns a balance from a currency from a specific account at a specific exchange.
func (b *BalancesStrategy) Currency(exchangeName string, code string, accountID string) (account.Balance, error) {
	holdings, loaded := b.Load(exchangeName)
	if !loaded {
		var empty account.Balance
		return empty, ErrHoldingsNotFound
	}

	for _, sub := range holdings.Accounts {
		if sub.ID == accountID {
			for _, balance := range sub.Currencies {
				if balance.CurrencyName.String() == code {
					return balance, nil
				}
			}
		}
	}
	return account.Balance{}, ErrCurrencyNotFound
}

// +--------------------+
// | Strategy interface |
// +--------------------+

// Init is called when the strategy is first initialized. It takes in a `Dealer` and an `exchange.IBotExchange` as parameters.\
// We get the refresh rate from `b`
func (b *BalancesStrategy) Init(d *Dealer, e exchange.IBotExchange) error {
	return b.ticker.Init(d, e)
}

// OnFunding is called when a funding event occurs. The `FundingData` struct contains the funding data.
// Updates the holdings of the account. It basically looks at the side of the OCO(1), then the price of the oco the amount is either reduced by the FundingRate
// which means it has increased our balance, either by getting removed altogether, which means is has decreased our balance.
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
	return b.ticker.Init(d, e)
}
