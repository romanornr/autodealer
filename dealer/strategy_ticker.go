package dealer

import (
	"github.com/romanornr/autodealer/util"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"sync"
	"time"
)

// TickerStrategy is a struct that implements the TickerStrategy interface.
// Interval is the time interval between tickers.
// TickerFunc is a function that returns a ticker.
// tickers is a map of tickers.
type TickerStrategy struct {
	Interval time.Duration
	TickerFunc func(d *Dealer, e exchange.IBotExchange)
	tickers sync.Map
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

func (s *TickerStrategy) OnFunding(d *Dealer, e exchange.IBotExchange, x stream.FundingData) error {
	return nil
}