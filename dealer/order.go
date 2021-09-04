package dealer

import (
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"sync"
	"sync/atomic"
)

type OrderKey struct {
	ExchangeName string
	OrderID string
}

type OrderValue struct {
	SubmitResponse order.SubmitResponse
	userData interface{}
 }

//type OrderRegistry struct {
//	length int32
//	values sync.Map
//}

type OrderRegistry struct {
	length int32
	values map[string]*OrderKey
	Mutex sync.RWMutex
}

func NewOrderRegistry() *OrderRegistry {
	return &OrderRegistry{
		length: 0,
		values: make(map[string]*OrderKey),
		Mutex:  sync.RWMutex{},
	}
}

func (r *OrderRegistry) Store(exchangeName string, response order.SubmitResponse, userdata interface{}) bool {
	key := OrderKey{
		ExchangeName: exchangeName,
		OrderID:      response.OrderID,
	}
	r.Mutex.Lock()
	if _, loaded := r.values[key.OrderID]; !loaded {
		r.values[key.OrderID] = &key
		r.Mutex.Unlock()
		atomic.AddInt32(&r.length, 1)
		return true
	}
	r.Mutex.Unlock()
	return false
}

func (r *OrderRegistry) GetOrderValue(exchangeName, orderID string) (OrderValue, bool) {
	key := OrderKey{
		ExchangeName: exchangeName,
		OrderID:      orderID,
	}
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()
	val, ok := r.values[key.OrderID]
	if !ok {
		return OrderValue{}, ok
	}

	return OrderValue{
		SubmitResponse: order.SubmitResponse{},
		userData:       val.UserData,
 	}, true
}



