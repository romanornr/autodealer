package dealer

import (
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

// OrderKey struct implements the `Key` interface of the sync.Map, which used for type assertion of the key
type OrderKey struct {
	ExchangeName string
	OrderID      string
}

// OrderValue struct holds two fields, which are both stored under the `OrderValue` struct.
// The first field is the `SubmitResponse` which comes from `Submit` function. It's the response returned from the webserver of each exchange.
// It includes data like whether or not the order is placed, the order ID (in case the order is placed), the creation timestamp, the creation amount (unit), etc.
// It also contains all exchange-specific error messages under `Error()`.
type OrderValue struct {
	SubmitResponse order.SubmitResponse
	UserData       interface{}
}

// This code begins by atomically reading the length property.
// This indicates that it loads the current atomic value of the int property defined as an anonymous field and exported through atomic.AddInt32(&r.length, 1).
// Why are we doing this? If we attempted to change an exported field, we might wind up with a scenario in which both the outer and inner loops attempting
// to modify and access the length property at the same time affect the same variable.

// OrderRegistry struct that stores two things: The amount of orders in it and the key.
// It has a single `int` property that represents the amount of orders currently in the registry.
// The `int` property is an atomic.Int, which is part of the golang's atomic package.
// It contains the modification of this `int`property that happens at the same time. The modification of an int cannot happen at two places in code in parallel. This provides a safe way to get or update integers from multiple routines, or from goroutines. In this case it’s the inner length field.
type OrderRegistry struct {
	length int32
	values sync.Map
}

// NewOrderRegistry constructs a new OrderRegistry. The function initializes the field atomic.Int32 called length with 0
// this means your r.length is incremented after every call of this function.
func NewOrderRegistry() *OrderRegistry {
	return &OrderRegistry{
		length: 0,
		values: sync.Map{},
	}
}

// Store stores the global order data to the `OrderRegistry`. The code looks like it should just wrap structs around another struct, that is not the case.
// First, it checks if an order with the same exchange name and order ID already exist, if not the order value will be stored in the `OrderRegistry` and `returned` will `true`.
// Secondly, the `loaded` argument is returned. This `loaded` argument is needed to see if the order has already been added to the `OrderRegistry` and therefore we need to run the code again.
func (r *OrderRegistry) Store(exchangeName string, response order.SubmitResponse, userData interface{}) bool {
	key := OrderKey{
		ExchangeName: exchangeName,
		OrderID:      response.OrderID,
	}
	value := OrderValue{
		SubmitResponse: response,
		UserData:       userData,
	}
	_, loaded := r.values.LoadOrStore(key, value)

	if !loaded {
		// If not loaded, then it's stored, so length++.
		atomic.AddInt32(&r.length, 1)
	}

	return !loaded
}

// GetOrderValue initially verifies that the exchange name and order ID exist in the OrderRegistry.
// This results in the `want`, and `ok` variables and assignment then return if that's the case.
// If it's not, then it's safe for r.values. LoadOrStore to return a "new" order value.
// It will create a new OrderValue with the input `UserData` and then load the `SubmitResponse` from the `DecodedUpdates` channel. If that should fail, its logged
func (r *OrderRegistry) GetOrderValue(exchangeName, orderID string) (OrderValue, bool) {
	key := OrderKey{
		ExchangeName: exchangeName,
		OrderID:      orderID,
	}

	var (
		loaded bool
		ok   bool
		pointer  interface{}
		value OrderValue
	)

	if pointer, loaded = r.values.Load(key); loaded {
		value, ok = pointer.(OrderValue)
		if !ok {
			logrus.Fatalf("have %T, want OrderValue", pointer)
		}
	}

	return value, ok
}

func (r *OrderRegistry) Length() int {
	return int(atomic.LoadInt32(&r.length))
}
