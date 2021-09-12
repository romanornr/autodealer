package dealer

import (
	"errors"
	"testing"
)

func TestNewRootStrategy(t *testing.T) {
	s := NewRootStrategy()
	s.Add("test", &s)

	if _, err := s.Get("test"); err != nil {
		t.Fatalf("expected err to be nil, got %v", err)
	}

	if _, err := s.Get("error"); err != ErrStrategyNotFound {
		t.Fatalf("expected err to be %v, got %v", ErrStrategyNotFound, err)
	}

	if _, err := s.Get(""); err != ErrStrategyNotFound {
		t.Fatalf("expected err to be %v, got %v", ErrStrategyNotFound, err)
	}

	a, err := s.Delete("test")
	if err != nil {
		t.Fatalf("expected err to be nil, got %v", err)
	}

	if a == nil {
		t.Fatalf("expected err to be nil")
	}

	a, err = s.Delete("test")
	if err != ErrStrategyNotFound {
		t.Fatalf("expected err to be %v, got %v", ErrStrategyNotFound, err)
	}
	if a != nil {
		t.Fatalf("expected a to be nil, got %v", a)
	}

	a, err = s.Delete("")
	if err != ErrStrategyNotFound {
		t.Fatalf("expected err to be %v, got %v", ErrStrategyNotFound, err)
	}
	if a != nil {
		t.Fatalf("expected a to be nil, got %v", a)
	}
}

func TestRootStrategyEach(t *testing.T) {
	s := NewRootStrategy()
	s.Add("test", &s)
	s.Add("test2", &s)
	if err := s.each(func(Strategy) error { return nil }); err != nil {
		t.Fatalf("expected err to be %v, got %v", nil, err)
	}
	if err := s.each(func(Strategy) error { return errors.New("error") }); err == nil {
		t.Fatalf("expected err to be %v, got %v", ErrStrategyNotFound, nil)
	}
}

//func TestNewRootStrategyInit(t *testing.T) {
//	s := NewRootStrategy()
//	s.Add("test", &s)
//	d, err := NewBuilder().Build()
//	if err != nil {
//		t.Fatalf("expected err to be %v, got %v", nil, err)
//	}
//
//	factory := make(ExchangeFactory)
//	testExchange := ftx.FTX{}
//	factory.Register("ftx", func() (exchange.IBotExchange, error) {
//		return &testExchange, nil
//	})
//
//	exchangeFactory, err := factory.NewExchangeByName("ftx")
//	if err != nil {
//		t.Fatalf("failed to create exchange factory: %v\n", err)
//	}
//
//	err = s.Init(d, exchangeFactory)
//	if err != nil {
//		t.Fatalf("expected err to be %v, got %v", nil, err)
//	}
//
//}

//func TestNewRootStrategyInit(t *testing.T) {
//	s := NewRootStrategy()
//	s.Add("test", &s)
//	d, err := NewBuilder().Build()
//	if err != nil {
//		t.Fatalf("expected err to be %v, got %v", nil, err)
//	}
//
//	if err = s.Init(d, nil); err != nil {
//		t.Fatal(err)
//	}

//if err = s.Init(d, nil); err != ErrStrategyNotFound {
//	t.Fatalf("expected err to be %v, got %v", ErrNotStrategy, err)
//}
//}
