package orderbuilder

import (
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

// OrderBuilder is a struct that holds the order details
type OrderBuilder struct {
	Order order.Submit
}

////BuildProcess builds the order
//type BuildProcess interface {
//	AtExchange(string) BuildProcess
//	ForAccountID(string) BuildProcess
//	ForCurrencyPair(currency.Pair) BuildProcess
//	WithAssetType(item asset.Item) BuildProcess
//	ForPrice(float64) BuildProcess
//	WithAmount(float64) BuildProcess
//	UseOrderType(order.Type) BuildProcess
//	SetSide(order.Side) BuildProcess
//	Build() (*order.Submit, error)
//}
//
//// Director is the interface for the order builder
//type Director struct {
//	builder BuildProcess
//}
//
//// SetBuilder is the method for setting the order builder
//func (d *Director) SetBuilder(b BuildProcess) {
//	d.builder = b
//}
//
//// Construct is the method for building the order
//func (d Director) Construct() (*order.Submit, error) {
//	return d.builder.Build()
//}
//
//// NewOrderBuilder returns a new instance of OrderBuilder
//func NewOrderBuilder() *OrderBuilder {
//	return &OrderBuilder{Order: order.Submit{}}
//}
//
//// AtExchange sets the exchange for the order
//func (ob *OrderBuilder) AtExchange(exchange string) BuildProcess {
//	ob.Order.Exchange = exchange
//	return ob
//}
//
//// ForCurrencyPair sets the currency pair for the order
//func (ob *OrderBuilder) ForCurrencyPair(pair currency.Pair) BuildProcess {
//	ob.Order.Pair = pair
//	return ob
//}
//
//// ForAccountID sets the account ID for the order
//func (ob *OrderBuilder) ForAccountID(accountID string) BuildProcess {
//	ob.Order.AccountID = accountID
//	return ob
//}
//
//// WithAssetType sets the asset type for the order
//func (ob *OrderBuilder) WithAssetType(item asset.Item) BuildProcess {
//	ob.Order.AssetType = item
//	return ob
//}
//
//// ForPrice sets the price for the order
//func (ob *OrderBuilder) ForPrice(price float64) BuildProcess {
//	ob.Order.Price = price
//	return ob
//}
//
//// WithAmount sets the amount for the order
//func (ob *OrderBuilder) WithAmount(amount float64) BuildProcess {
//	ob.Order.Amount = amount
//	return ob
//}
//
//// UseOrderType sets the order type for the order
//func (ob *OrderBuilder) UseOrderType(orderType order.Type) BuildProcess {
//	ob.Order.Type = orderType
//	return ob
//}
//
//// SetSide sets the side for the order
//func (ob *OrderBuilder) SetSide(side order.Side) BuildProcess {
//	ob.Order.Side = side
//	return ob
//}
//
//// Build returns the order
//func (ob *OrderBuilder) Build() (*order.Submit, error) {
//	if err := ob.Order.Validate(); err != nil {
//		return &order.Submit{}, err
//	}
//	return &ob.Order, nil
//}

// NewOrderBuilder returns a new instance of OrderBuilder
func NewOrderBuilder() *OrderBuilder {
	return &OrderBuilder{Order: order.Submit{}}
}

func (o *OrderBuilder) AtExchange(exchange string) *OrderBuilder {
	o.Order.Exchange = exchange
	return o
}

func (o *OrderBuilder) ForCurrencyPair(pair currency.Pair) *OrderBuilder {
	o.Order.Pair = pair
	return o
}

func (o *OrderBuilder) WithAssetType(assetType asset.Item) *OrderBuilder {
	o.Order.AssetType = assetType
	return o
}

func (o *OrderBuilder) ForPrice(price float64) *OrderBuilder {
	o.Order.Price = price
	return o
}

func (o *OrderBuilder) WithAmount(amount float64) *OrderBuilder {
	o.Order.Amount = amount
	return o
}

func (o *OrderBuilder) UseOrderType(orderType order.Type) *OrderBuilder {
	o.Order.Type = orderType
	if orderType == order.Limit {
		o.WithPostOnly(true)
	}
	return o
}

func (o *OrderBuilder) SetSide(side order.Side) *OrderBuilder {
	o.Order.Side = side
	return o
}

func (o *OrderBuilder) WithPostOnly(postOnly bool) *OrderBuilder {
	o.Order.PostOnly = postOnly
	return o
}

func (o *OrderBuilder) SetReduceOnly(reduce bool) *OrderBuilder {
	o.Order.ReduceOnly = reduce
	return o
}

func (o *OrderBuilder) UseImmediateOrCancel(immediateOrCancel bool) *OrderBuilder {
	o.Order.ImmediateOrCancel = immediateOrCancel
	return o
}

func (o *OrderBuilder) ForAccountID(accountID string) *OrderBuilder {
	o.Order.AccountID = accountID
	return o
}

func (o *OrderBuilder) Build() (*order.Submit, error) {
	if err := o.Order.Validate(); err != nil {
		return &order.Submit{}, err
	}
	return &o.Order, nil
}
