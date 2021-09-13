package dealer

import (
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ftx"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"testing"
	"time"
)

func TestTickerStrategy(t *testing.T) {
	d, err := NewBuilder().Build()
	if err != nil {
		t.Fatalf("expected err to be %v, got %v", nil, err)
	}

	factory := make(ExchangeFactory)
	testExchange := ftx.FTX{}
	factory.Register("ftx", func() (exchange.IBotExchange, error) {
		return &testExchange, nil
	})

	exchangeFactory, err := factory.NewExchangeByName("ftx")
	if err != nil {
		t.Fatalf("failed to create exchange factory: %v\n", err)
	}

	strategy := TickerStrategy{
		Interval: time.Millisecond * 5,
		TickerFunc: func(d *Dealer, e exchange.IBotExchange) {
			t.Log("failed")
		},
	}
	err = strategy.Init(d, exchangeFactory)
	if err != nil {
		t.Failed()
	}
}

func TestTickerStrategyOnFunding(t *testing.T) {
	d, err := NewBuilder().Build()
	if err != nil {
		t.Fatalf("expected err to be %v, got %v", nil, err)
	}

	factory := make(ExchangeFactory)
	testExchange := ftx.FTX{}
	factory.Register("ftx", func() (exchange.IBotExchange, error) {
		return &testExchange, nil
	})

	exchangeFactory, err := factory.NewExchangeByName("ftx")
	if err != nil {
		t.Fatalf("failed to create exchange factory: %v\n", err)
	}

	tickerStrategy := TickerStrategy{
		Interval: time.Millisecond * 5,
		TickerFunc: func(d *Dealer, e exchange.IBotExchange) {
			t.Log("failed barf")
		},
	}

	err = tickerStrategy.Init(d, exchangeFactory)
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

	err = tickerStrategy.OnFunding(d, exchangeFactory, data)
	if err != nil {
		t.Fatalf("expected err to be %v, got %v", nil, err)
	}
}

func TestDeleteTickerFunc(t *testing.T) {
	d, err := NewBuilder().Build()
	if err != nil {
		t.Fatalf("expected err to be %v, got %v", nil, err)
	}

	factory := make(ExchangeFactory)
	testExchange := ftx.FTX{}
	factory.Register("ftx", func() (exchange.IBotExchange, error) {
		return &testExchange, nil
	})

	exchangeFactory, err := factory.NewExchangeByName("ftx")
	if err != nil {
		t.Fatalf("failed to create exchange factory: %v\n", err)
	}

	tickerStrategy := TickerStrategy{
		Interval: time.Millisecond * 5,
		TickerFunc: func(d *Dealer, e exchange.IBotExchange) {
			t.Log("failed barf")
		},
	}

	err = tickerStrategy.Init(d, exchangeFactory)
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

	if err = tickerStrategy.OnFunding(d, exchangeFactory, data); err != nil{
		t.Fatalf("expected err to be %v, got %v", nil, err)
	}
}