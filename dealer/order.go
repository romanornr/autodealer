package dealer

import (
	"github.com/sirupsen/logrus"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"sync"
	"sync/atomic"
)

type OrderKey struct {
	ExchangeName string
	OrderID      string
}

type OrderValue struct {
	SubmitResponse order.SubmitResponse
	UserData       interface{}
}

type OrderRegistry struct {
	length int32
	values sync.Map
	Mutex  sync.RWMutex
}

func NewOrderRegistry() *OrderRegistry {
	return &OrderRegistry{
		length: 0,
		values: sync.Map{},
		Mutex:  sync.RWMutex{},
	}
}

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
