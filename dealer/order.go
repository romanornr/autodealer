package dealer

import (
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

type OrderKey struct {
	ExchangeName string
	OrderID      string
}

type OrderValue struct {
	SubmitResponse order.SubmitResponse
	UserData       interface{}
}

// An OrderRegistry is a struct for keeping track of bookings; it has two
// exported properties: 'length' and 'values'.
// OrderRegistry is safe for concurrent use by multiple goroutines.
type OrderRegistry struct {
	length int32
	values sync.Map
	Mutex  sync.RWMutex
}

// NewOrderRegistry creates a new OrderRegistry
// Short is unique per unique song id. This is the only value stored, so it is guaranteed to be unique
func NewOrderRegistry() *OrderRegistry {
	return &OrderRegistry{
		length: 0,
		values: sync.Map{},
		Mutex:  sync.RWMutex{},
	}
}

// Store Stores an order.submit response in the order registry. If the response already exists in
// the order registry it is not stored again. Returns `true` if an order value was stored in
// the registry and `false` otherwise.
func (r *OrderRegistry) Store(exchangeName string, response order.SubmitResponse, userData interface{}) bool {
	key := OrderKey{
		ExchangeName: exchangeName,
		OrderID:      response.OrderID,
	}
	value := OrderValue{
		SubmitResponse: response,
		UserData:       userData,
	}

	if _, loaded := r.values.LoadOrStore(key, value); !loaded {
		atomic.AddInt32(&r.length, 1)
		return loaded
	}
	return false
}

// GetOrderValue returns the order value for the given exchange name and
// order ID, or false if it's not found. Sees to that the type assertion is valid.
func (r *OrderRegistry) GetOrderValue(exchangeName, orderID string) (OrderValue, bool) {
	key := OrderKey{
		ExchangeName: exchangeName,
		OrderID:      orderID,
	}

	var (
		ok   bool
		val  interface{}
		want OrderValue
	)

	val, ok = r.values.Load(key)

	if ok {
		want, ok = val.(OrderValue)
		if !ok {
			logrus.Fatalf("have %T, want OrderValue", val)
		}
	}
	return want, ok
}
