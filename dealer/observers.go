package dealer

import (
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

// OnFilledObserver
// An observer that responds to each placed order by the dealer
type OnFilledObserver interface {
	OnFilled(d *Dealer, e exchange.IBotExchange, x order.Detail)
}

// Slots have an OnFilled function pointer.
type Slots struct {
	OnFilledSlot func(d *Dealer, e exchange.IBotExchange, x order.Detail)
}

// OnFilled is invoked with the placed order and the exchange
// per Observer, we can see we can retrieve its original details and possibly use
// it as a source of data and reconcile using this. If we agreed per exchange this strategy is acceptable, all orders are needed.
func (s Slots) OnFilled(d *Dealer, e exchange.IBotExchange, x order.Detail) {
	if s.OnFilledSlot != nil {
		s.OnFilledSlot(d, e, x)
	}
}
