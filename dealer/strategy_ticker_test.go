package dealer

import (
	"testing"
	"time"

	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
)

func TestTickerStrategy(t *testing.T) {
	d, err := NewBuilder().Build()
	if err != nil {
		t.Fatalf("expected err to be %v, got %v", nil, err)
	}

	e, err := d.ExchangeManager.GetExchangeByName("FTX")
	if err != nil {
		t.Errorf("expected no error, got %s\n", err)
	}
	strategy := TickerStrategy{
		Interval: time.Second * 1,
		TickFunc: func(d *Dealer, e exchange.IBotExchange) {
			t.Log("failed")
		},
	}
	err = strategy.Init(d, e)
	if err != nil {
		t.Failed()
	}
}

func TestTickerStrategyOnFunding(t *testing.T) {
	d, err := NewBuilder().Build()
	if err != nil {
		t.Fatalf("expected err to be %v, got %v", nil, err)
	}

	e, err := d.ExchangeManager.GetExchangeByName("FTX")
	if err != nil {
		t.Errorf("expected no error, got %s\n", err)
	}

	tickerStrategy := TickerStrategy{
		Interval: time.Second * 1,
		TickFunc: func(d *Dealer, e exchange.IBotExchange) {
			t.Log("failed barf")
		},
	}

	err = tickerStrategy.Init(d, e)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	data := stream.FundingData{
		Timestamp:    time.Now(),
		CurrencyPair: currency.NewPair(currency.BTC, currency.USDT),
		AssetType:    asset.Spot,
		Exchange:     "",
		Amount:       0,
		Rate:         0,
		Period:       0,
		Side:         order.AnySide,
	}

	err = tickerStrategy.OnFunding(d, e, data)
	if err != nil {
		t.Fatalf("expected err to be %v, got %v", nil, err)
	}
}

func TestDeleteTickerFunc(t *testing.T) {
	d, err := NewBuilder().Build()
	if err != nil {
		t.Fatalf("expected err to be %v, got %v", nil, err)
	}

	e, err := d.ExchangeManager.GetExchangeByName("FTX")
	if err != nil {
		t.Errorf("expected no error, got %s\n", err)
	}

	tickerStrategy := TickerStrategy{
		Interval: time.Millisecond * 5,
		TickFunc: func(d *Dealer, e exchange.IBotExchange) {
			t.Log("failed barf")
		},
	}

	err = tickerStrategy.Init(d, e)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	data := stream.FundingData{
		Timestamp:    time.Now(),
		CurrencyPair: currency.NewPair(currency.BTC, currency.USDT),
		AssetType:    asset.Spot,
		Exchange:     "",
		Amount:       0,
		Rate:         0,
		Period:       0,
		Side:         order.AnySide,
	}

	if err = tickerStrategy.OnFunding(d, e, data); err != nil {
		t.Fatalf("expected err to be %v, got %v", nil, err)
	}
}
