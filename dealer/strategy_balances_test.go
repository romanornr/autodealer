package dealer

import (
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"sync"
	"testing"
	"time"
)

func TestNewBalancesStrategy(t *testing.T) {
	b := NewBalancesStrategy(100*time.Millisecond)
	if b == nil {
		t.Logf("expected balances strategy to not nil")
		t.Failed()
	}
}

func TestNewBalancesStrategyWithError(t *testing.T) {
	b := NewBalancesStrategy(1 * time.Second)
	if b != nil {
		t.Logf("expected not nil")
		t.Failed()
	}
}

func TestInitBalancesStrategy(t *testing.T) {

	ts := TickerStrategy{
		Interval: time.Second * 1,
		TickFunc: func(d *Dealer, e exchange.IBotExchange) {
			t.Log("failed")
		},
	}

	b := &BalancesStrategy{
		balances: sync.Map{},
		ticker:   ts,
	}
	d, err := NewBuilder().Build()
	if err != nil {
		t.Errorf("expected no error, got %v\n", err)
	}
	e, err := d.ExchangeManager.GetExchangeByName("ftx")
	if err != nil {
		t.Errorf("expected error, got %s\n", err)
	}

	if err = b.Init(d, e); err != nil {
		t.Errorf("expected no error, got %v",err)
	}
}

func TestInitBalancesStrategyError(t *testing.T) {
	ts := TickerStrategy{
		Interval: time.Second * 1,
		TickFunc: func(d *Dealer, e exchange.IBotExchange) {
			t.Log("failed")
		},
	}

	b := &BalancesStrategy{
		balances: sync.Map{},
		ticker:   ts,
	}

	d, err := NewBuilder().Build()
	if err != nil {
		t.Errorf("expected no error, got %v\n", err)
	}
	e, err := d.ExchangeManager.GetExchangeByName("ftx")
	if err != nil {
		t.Errorf("expected error, got %s\n", err)
	}
	if err = b.Init(d, e); err != nil {
		t.Errorf("expected no error, got %v\n", err)
	}
}