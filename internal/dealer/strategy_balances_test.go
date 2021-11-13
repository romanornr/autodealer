package dealer

import (
	"context"
	"github.com/sirupsen/logrus"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"sync"
	"testing"
	"time"
)

func TestNewBalancesStrategy(t *testing.T) {
	b := NewBalancesStrategy(100 * time.Millisecond)
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

	if err = b.Init(context.Background(), d, e); err != nil {
		t.Errorf("expected no error, got %v", err)
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
	if err = b.Init(context.Background(), d, e); err != nil {
		t.Errorf("expected no error, got %v\n", err)
	}
}

func TestStoreBalancesStrategyError(t *testing.T) {
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
	if err = b.Init(context.Background(), d, e); err != nil {
		t.Errorf("expected no error, got %v\n", err)
	}

	a, err := e.FetchAccountInfo(context.Background(), asset.Spot)
	if err != nil {
		logrus.Errorf("fetch account info failed: %v\n", err)
	}

	b.Store(a)

	holdings, ok := b.balances.Load(e.GetName())
	if !ok {
		t.Errorf("expected no error, got %v\n", err)
	}

	x := holdings.(account.Holdings)

	if holdings == nil {
		t.Errorf("expected stored holding %s to be not nil, got nil\n", e.GetName())
	}

	if len(x.Accounts) > 0 {
		t.Errorf("expected account count to be > 0, got %d\n", len(x.Accounts))
	}
}
