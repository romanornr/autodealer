package dealer

import (
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
