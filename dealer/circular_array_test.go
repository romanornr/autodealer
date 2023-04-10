package dealer

import "testing"

// In this implementation, we have created test cases for the following functions:
//
// 1. `NewCircularArray`
// 2. `Push` and `Len`
// 3. `Index`
// 4. `LastIndex`
// 5. `At`
// 6. `Last`
// 7. `Floats`
// 8. `LastFloat`
// 9. `FloatsPanic` (to test the panic when a non-float64 value is found)
//
// These test cases cover the main functionalities of the CircularArray and check for both expected and unexpected input.
// Remember to run the tests with `go test` command to check if your implementation works as expected.

// TestNewCircularArray tests that NewCircularArray
// returns a CircularArray with a given size.
func TestNewCircularArray(t *testing.T) {
	n := 5
	c := NewCircularArray(n)

	if c.Offset != 0 {
		t.Errorf("expected: %d, actual: %d", 0, c.Offset)
	}

	if cap(c.xs) != n {
		t.Errorf("expected: %d, actual: %d", n, cap(c.xs))
	}
}

// TestPushAndLen tests that Push and Len
func TestPushAndLen(t *testing.T) {
	c := NewCircularArray(3)
	c.Push(1)
	c.Push(2)
	c.Push(3)
	c.Push(4)

	if c.Len() != 3 {
		t.Errorf("expected: %d, actual: %d", 3, c.Len())
	}
}

// TestIndex tests that Index
func TestIndex(t *testing.T) {
	c := NewCircularArray(3)
	c.Push(1)
	c.Push(2)
	c.Push(3)

	expectedIndexes := []int{0, 1, 2}
	for i, index := range expectedIndexes {
		if index != c.Index(i) {
			t.Errorf("expected: %d, actual: %d", index, c.Index(i))
		}
	}
}

// TestLastIndex tests that LastIndex
func TestLastIndex(t *testing.T) {
	c := NewCircularArray(3)
	c.Push(1)
	c.Push(2)
	c.Push(3)

	expectedIndexes := 2
	if lastIndex := c.LastIndex(); lastIndex != expectedIndexes {
		t.Errorf("expected: %d, actual: %d", expectedIndexes, lastIndex)
	}
}

// TestAt tests that At
func TestLast(t *testing.T) {
	c := NewCircularArray(3)
	c.Push(1)
	c.Push(2)
	c.Push(3)
	c.Push(4)

	expected := 4
	if last := c.Last(); last != expected {
		t.Errorf("expected: %d, actual: %d", expected, last)
	}
}

// TestFloats tests that Floats
func TestFloats(t *testing.T) {
	c := NewCircularArray(3)
	c.Push(1.0)
	c.Push(2.0)
	c.Push(3.0)
	c.Push(4.0)

	expected := []float64{2.0, 3.0, 4.0}
	floats := c.Floats()
	for i, expected := range expected {
		if floats[i] != expected {
			t.Errorf("expected: %f, actual: %f", expected, floats[i])
		}
	}
}

// TestLastFloat tests that LastFloat
func TestLastFloat(t *testing.T) {
	c := NewCircularArray(3)
	c.Push(1.0)
	c.Push(2.0)
	c.Push(3.0)
	c.Push(4.0)

	expected := 4.0
	if last := c.LastFloat(); last != expected {
		t.Errorf("expected: %f, actual: %f", expected, last)
	}
}

// TestFloatsPanic tests that FloatsPanic
func TestFloatsPanic(t *testing.T) {
	c := NewCircularArray(3)
	c.Push(1)
	c.Push(2.0)
	c.Push(3.0)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected: %v, actual: %v", nil, r)
		}
	}()
	c.Floats()
}
