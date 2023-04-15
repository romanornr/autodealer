package dealer

import "fmt"

// CircularArray is initialized with the "width" of the array.
type CircularArray struct {
	Offset int
	xs     []interface{}
}

// NewCircularArray function creates a new circular buffer of a defined size of capacity.
// The Len function returns the length of the buffer which is the same as the number of observations in a specific window.
func NewCircularArray(n int) CircularArray {
	xs := make([]interface{}, 0, n)
	if cap(xs) != n {
		panic("")
	}

	return CircularArray{
		Offset: 0,
		xs:     xs,
	}
}

// Index maps an external 0-based index to the corresponding internal index.
func (a *CircularArray) Index(i int) int {
	return (i + a.Offset) % cap(a.xs)
}

// The LastIndex function returns int value. If the values are all inside of uint values (whose maximum value is (1 << 31) - 1) the must value of offset, which is returned by lastIndex function, would be (1 << 31) - 1. In normal case, the maximum uint value will be (1 << 32) - 1. Then, lastIndex function returns (1 << 32) - 1.
func (a *CircularArray) LastIndex() int {
	return a.Index(len(a.xs) - 1)
}

func (a *CircularArray) Push(x interface{}) {
	if len(a.xs) < cap(a.xs) {
		a.xs = append(a.xs, x)
	} else {
		a.Offset = (a.Offset + 1) % cap(a.xs)
		last := a.LastIndex()
		a.xs[last] = x
	}
}

// +-----------------+
// | Array interface |
// +-----------------+

// Len function returns the number of elements currently stored. It returns `len(y)` which extracts the length of the underlying slice.
func (a *CircularArray) Len() int {
	return len(a.xs)
}

// At function returns the n-th element which can be indexed by the external `index` argument.
// To map this external 0-based index to the underlying 0-based index of the slice, we call a helper method `Index` which will return a wrapped index.
// Assuming a CircularArray of 100 elements, a call to `At(99)` with a wrapping index of 96 will return the 96th element of the array, wrapped completely around the array as considered as a circle.
func (a *CircularArray) At(index int) interface{} {
	mapped := a.Index(index)
	return a.xs[mapped]
}

// Last returns the last element of the underlying slice.
func (a *CircularArray) Last() interface{} {
	return a.At(a.Len() - 1)
}

// Floats converts each element in the underlying slice to a floating point number.
func (a *CircularArray) Floats() []float64 {
	ys := make([]float64, a.Len())

	for i := 0; i < a.Len(); i++ {
		x := a.At(i)
		if y, ok := x.(float64); !ok {
			panic(fmt.Sprintf("illegal type: %s\n", x))
		} else {
			ys[i] = y
		}
	}
	return ys
}

// LastFloat function returns the last element in the underlying slice cast to a floating point number
func (a *CircularArray) LastFloat() float64 {
	return a.Last().(float64)
}
