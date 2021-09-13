package dealer

import (
	"github.com/romanornr/autodealer/util"
	"github.com/sirupsen/logrus"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"sync"
	"time"
)

// tickers are meant to store an interval time once they are registered in the Init function.
// Once you remove the ticker from the tickers map, you are implicitly stopping the ticker.
// You can't stop it explicitly because you have no idea when will the Init return.
// The ticker is being removed when a new ticker replaces a previous ticker.
//
// When a user calls to init a ticker, a ticker is stored in a map with a key of the exchange type, `pointer.LoadOrStore`.
// Finally, to get a pointer to a stored ticker, you have to access it through a pointer type. since a map is not a native.
//
// You might be wondering, why did I name the tickers map `pointer`?
// So far, the only place that is using pointers is when fetching a ticker from a `pointer.LoadOrStore`,
// which is when we store a ticker in the tickers map.
//
// At the end of deinit, we remove the ticker and the time in the store is lost.

// TickerStrategy is a struct that implements the TickerStrategy interface.
// Interval is the time interval between tickers.
// TickerFunc is a function that returns a ticker.
// tickers is a map of tickers.
type TickerStrategy struct {
	Interval   time.Duration
	TickerFunc func(d *Dealer, e exchange.IBotExchange)
	tickers    sync.Map
}

// Init starts the ticker for the given exchange
//
// Parameters:
//     d *Dealer - the dealer object
//     e exchange.IBotExchange - the exchange to start the ticker for
//
// Returns:
//     error - any errors that occurred
func (s *TickerStrategy) Init(d *Dealer, e exchange.IBotExchange) error {
	ticker := *time.NewTicker(s.Interval)

	if s.TickerFunc != nil {
		go func() {
			util.CheckerPush()
			defer util.CheckerPop()

			// Call now initially
			s.TickerFunc(d, e)
			for range ticker.C {
				s.TickerFunc(d, e)
			}
		}()
	}
	_, loaded := s.tickers.LoadOrStore(e.GetName(), ticker)
	if loaded {
		panic("one exchange can have just one ticker")
	}
	return nil
}

// OnFunding - entrypoint for the strategy triggered on funding data
// Called if OnMsgHandlerMsg(Data(fundingData) trigger returns true
func (s *TickerStrategy) OnFunding(d *Dealer, e exchange.IBotExchange, x stream.FundingData) error {
	return nil
}

// OnPrice is called when price announcement or updates are received.
func (s *TickerStrategy) OnPrice(d *Dealer, e exchange.IBotExchange, x ticker.Price) error {
	return nil
}

// OnKline is called when tick data is received
func (s *TickerStrategy) OnKline(d *Dealer, e exchange.IBotExchange, x stream.KlineData) error {
	return nil
}

// OnOrderBook is triggered when API receives orderbook update
func (s *TickerStrategy) OnOrderBook(d *Dealer, e exchange.IBotExchange, x orderbook.Base) error {
	return nil
}

// OnOrder manipulates the dealers orderbook given a new incoming order
// x order.Detail is the new order
func (s *TickerStrategy) OnOrder(d *Dealer, e exchange.IBotExchange, x order.Detail) error {
	return nil
}

// OnModify updates modify order
func (s *TickerStrategy) OnModify(d *Dealer, e exchange.IBotExchange, x order.Modify) error {
	return nil
}

// OnBalanceChange implements the TickerStrategy interface function.
// Used to trigger an action for an already running ticker
func (s *TickerStrategy) OnBalanceChange(d *Dealer, e exchange.IBotExchange, x account.Change) error {
	return nil
}

// OnUnrecognized is a function that is called when an unsupported data type is given
// and most likely indicates a parsing error has occurred with Exchanges
func (s *TickerStrategy) OnUnrecognized(d *Dealer, e exchange.IBotExchange, x interface{}) error {
	return nil
}

// Deinit stops the ticker for the given Exchange.
func (s *TickerStrategy) Deinit(d *Dealer, e exchange.IBotExchange) {
	pointer, loaded := s.tickers.LoadAndDelete(e.GetName())
	if !loaded {
		panic("exchange has not registered a ticker")
	}

	tickers, ok := pointer.(time.Ticker)
	if !ok {
		logrus.Panicf("want time.Ticker, got %T", pointer)
	}
	tickers.Stop()
}
