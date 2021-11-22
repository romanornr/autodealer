package orderbuilder

import (
    "github.com/thrasher-corp/gocryptotrader/currency"
    "github.com/thrasher-corp/gocryptotrader/exchanges/asset"
    "github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

type OrderBuilder struct {
    Order order.Submit
}

// NewOrderBuilder returns a new instance of OrderBuilder
func NewOrderBuilder() *OrderBuilder {
    return &OrderBuilder{ Order: order.Submit{}}
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

