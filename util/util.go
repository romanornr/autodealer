package util

import (
	"runtime"
	"strings"
)

// Location attempts to write the name of the caller function's parent.
// This occurs when the pointer pc is set to 1 and when the compiler is queried for the function's name.
// The pointer's data type is set to the data type of the function that is currently being executed.
// The compiler is then queried to get the function's pointer. If it succeeds, the code then performs a location and completes the phrase
// If it cannot locate the function's pointer, it returns a question mark to indicate that it is unknown.
func Location() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "?"
	}
	fn := runtime.FuncForPC(pc)
	xs := strings.SplitAfterN(fn.Name(), "/", 3)
	//nolint: gomnd
	return xs[len(xs)-1]
 }

// Location2 implements the grandparent call interface
// and contains Call 'street calling' troubleshooting  and returns the name of the grandparent function
func Location2() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "?"
	}
	fn := runtime.FuncForPC(pc)
	xs := strings.SplitN(fn.Name(), "/", 3)
	return xs[len(xs)-1]
}

