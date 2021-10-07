package dealer

import (
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"strings"
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

	orderValue, _ := r.GetOrderValue("ftx", orderID)
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

func TestOrderRegistryDuplicate(t *testing.T) {
	orderID := "fake-order-id2"
	response := order.SubmitResponse{
		OrderID:       orderID,
		IsOrderPlaced: true,
	}

	r := NewOrderRegistry()
	duplicate := r.Store("ftx", response, nil)
	if duplicate {
		t.Failed()
	}

	if r.length != 1 {
		t.Errorf("Order Registry length count not incremented correctly")
		t.Failed()
	}

	duplicate2 := r.Store("ftx", response, nil)
	if duplicate2 == true {
		t.Logf("failed")
		t.Failed()
	}

	if r.length != 1 {
		t.Errorf("Order Registry length count not incremented correctly")
		t.Failed()
	}
}

func TestOrderRegistryGetExistingValue(t *testing.T) {
	r := NewOrderRegistry()
	response := order.SubmitResponse{
		OrderID:       "fake-order-id",
		IsOrderPlaced: true,
	}

	r.Store("ftx", response, nil)

	storedOrder, ok := r.GetOrderValue("ftx", "fake-order-id")
	if !ok {
		t.Error("order not found")
		t.Failed()
	}

	if response.IsOrderPlaced != storedOrder.SubmitResponse.IsOrderPlaced {
		t.Error("components mismatch")
		t.Failed()
	}
}

func TestOrderRegistryGetNonExistingValue(t *testing.T) {
	r := NewOrderRegistry()
	response := order.SubmitResponse{
		OrderID:       "fake-order-id",
		IsOrderPlaced: true,
	}

	r.Store("ftx", response, nil)
	if r.length != 1 {
		t.Error("Order Registry length count not incremented correctly")
		t.Failed()
	}

	storedOrder, ok := r.GetOrderValue("ftx", "fake-order-id")
	if !ok {
		t.Errorf("order not found")
		t.Failed()
	}

	if response.IsOrderPlaced != storedOrder.SubmitResponse.IsOrderPlaced {
		t.Error("Component mismatch")
		t.Failed()
	}

	storedOrder, ok = r.GetOrderValue("ftx", "fake-order-id-2")
	if ok {
		t.Error("order not found")
		t.Failed()
	}
}

func TestOrderRegistryContains(t *testing.T) {
	r := NewOrderRegistry()
	response := order.SubmitResponse{
		OrderID:       "fake-order-id",
		IsOrderPlaced: true,
	}

	r.Store("ftx", response, nil)

	if r.length != 1 {
		t.Error("Order Registry length count not incremented correctly")
		t.Failed()
	}

	val, _ := r.GetOrderValue("ftx", "fake-order-id")

	if !strings.Contains("fake-order-id", val.SubmitResponse.OrderID) {
		t.Error("order not found")
		t.Failed()
	}
}
