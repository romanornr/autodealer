package dealer

import (
	"context"
	"errors"

	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
)

var ErrNeedBalancesStrategy = errors.New("Dealer should be configured with balances support")

// Holdings function finds the exchanges balances strategy, then does two things: Looks for an exchange by its name in the holdings map.
// Returns the holding for the exchange if it's in the map.
func Holdings(d *Dealer, exchangeName string) (*ExchangeHoldings, error) {
	st, err := d.Root.Get("balances")
	if errors.Is(err, ErrStrategyNotFound) {
		return nil, ErrNeedBalancesStrategy
	}

	balances, ok := st.(*BalancesStrategy)
	if !ok {
		panic("cast failed")
	}

	holdings, err := balances.ExchangeHoldings(exchangeName)
	if err != nil {
		return nil, err
	}
	return holdings, nil
}

// ModifyOrder function will execute two steps: modify order on exchange current order status using the submitted ID, after that cancel that order using the same ID.
// All markers issue when modifying/canceling order also will be made using the same ID it was treated before
func ModifyOrder(ctx context.Context, d *Dealer, e exchange.IBotExchange, mod order.Modify) (ans order.ModifyResponse, err error) {
	ans, err = d.ModifyOrder(ctx, e, mod)
	if err == nil {
		return ans, nil
	}

	cancel := ModifyToCancel(mod)
	_ = d.CancelOrder(ctx, e, cancel)

	// Prepare submission
	var (
		submit   = ModifyToSubmit(mod)
		response *order.SubmitResponse
	)

	value, loaded := d.GetOrderValue(e.GetName(), mod.OrderID)
	if loaded {
		response, err = d.SubmitOrderUD(ctx, e, submit, value.UserData)
	} else {
		response, err = d.SubmitOrder(ctx, e, submit)
	}

	ans.Exchange = e.GetName()
	ans.AssetType = submit.AssetType
	ans.Pair = submit.Pair
	ans.OrderID = response.OrderID
	return ans, err
}

// Ticker takes the input and returns ticker price
func Ticker(p interface{}) ticker.Price {
	x, ok := p.(ticker.Price)
	if !ok {
		panic("")
	}
	return x
}

// ModifyToCancel function is responsible for creating a cancel order
func ModifyToCancel(mod order.Modify) order.Cancel {
	var cancel order.Cancel
	cancel.Exchange = mod.Exchange
	cancel.OrderID = mod.OrderID
	cancel.AssetType = mod.AssetType
	cancel.Pair = mod.Pair
	return cancel
}

// ModifyToSubmit function receives a Modify Order which is passed in from the client calls, it creates a Submit order from it.
// The client calls modifying the order and updating the fetched order if it's done. e.g if an order submitted is filled, the client should change to new order.
func ModifyToSubmit(mod order.Modify) order.Submit {
	sub := order.Submit{
		ImmediateOrCancel: mod.ImmediateOrCancel,
		//HiddenOrder:       mod.HiddenOrder,
		//FillOrKill:        mod.ImmediateOrCancel
		PostOnly:   mod.PostOnly,
		ReduceOnly: false, // Missing.
		//Leverage:          mod.Leverage,
		Price:  mod.Price,
		Amount: mod.Amount,
		//StopPrice:         0, // Missing.
		//LimitPriceUpper:   mod.LimitPriceUpper,
		//LimitPriceLower:   mod.LimitPriceLower,
		TriggerPrice: mod.TriggerPrice,
		//TargetAmount:      mod.TargetAmount,
		//ExecutedAmount:    mod.ExecutedAmount,
		//RemainingAmount:   mod.RemainingAmount,
		//Fee:               mod.Fee,
		Exchange: mod.Exchange,
		//InternalOrderID:   mod.InternalOrderID,
		//ID:                mod.ID,
		//AccountID:         mod.AccountID,
		//ClientID:          mod.ClientID,
		ClientOrderID: mod.ClientOrderID,
		//WalletAddress:     mod.WalletAddress,
		//Offset:            "", // Missing.
		Type: mod.Type,
		Side: mod.Side,
		//Status:            mod.Status,
		AssetType: mod.AssetType,
		//Date:              mod.Date,
		//LastUpdated:       mod.LastUpdated,
		Pair: mod.Pair,
		//Trades:            mod.Trades,
	}
	return sub
}
