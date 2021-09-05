package dealer

import (
	"fmt"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"testing"
)

func TestOrderRegistry(t *testing.T) {
	orderID := "fake-order-id"
	response := order.SubmitResponse{
		OrderID:       orderID,
		IsOrderPlaced: true,
	}
	r := NewOrderRegistry()
	r.Store("ftx", response, nil)

	orderValue, found := r.GetOrderValue("ftx", orderID)
	fmt.Println(found)
	if orderValue.SubmitResponse.OrderID != orderID {
		t.Logf("error should %s, but got %s\n", orderID, orderValue.SubmitResponse.OrderID)

	}

	if orderValue.UserData != orderID {
		t.Logf("error should %s, but got %s\n", orderID, orderValue.SubmitResponse.OrderID)

	}

	if r.length != 1 {
		t.Errorf("Order Registry length count not incremented correctly")
		t.Failed()
	}
}
