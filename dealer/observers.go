package dealer

import (
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

// A classic Observer design pattern implementation, allowing customizable post-order-fill actions.
// Slots refers to a function pointer. So the function can be called indirectly through the variable.
// The "OnFilledSlot" in the "Slots" struct is a function pointer, which is assigned a function that gets called when an order is filled.

// OnFilledObserver is an interface that responds to each placed order by the dealer.
// The OnFilled method is expected to perform operations when a trade order is filled.
type OnFilledObserver interface {
	OnFilled(d *Dealer, e exchange.IBotExchange, orderDetail order.Detail)
}

// Slots is a struct that contains an OnFilled function pointer.
// as the OnFilled method in the OnFilledObserver interface.
// This allows for customizable behavior when an order gets filled.
type Slots struct {
	OnFilledSlot func(d *Dealer, e exchange.IBotExchange, orderDetail order.Detail)
}

// OnFilled is invoked with the placed order and the exchange.
// It retrieves the original order details and can be used as a source of data for reconciliation.
// If this strategy is acceptable per exchange, all orders are needed.
func (s Slots) OnFilled(d *Dealer, e exchange.IBotExchange, orderDetail order.Detail) {
	if s.OnFilledSlot != nil {
		// Add error handling here to handle potential issues when calling s.OnFilledSlot
		s.OnFilledSlot(d, e, orderDetail)
	}
}
