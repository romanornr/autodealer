package dealer

import "testing"

func TestNewCircularArray(t *testing.T) {
	n := 10
	a := NewCircularArray(n)

	if a.Len() != 0 {
		t.Errorf("expected: %d, actual: %d", 0, a.Len())
	}

	if cap(a.xs) != n {
		t.Errorf("expected: %d, actual: %d", n, cap(a.xs))
	}
}
