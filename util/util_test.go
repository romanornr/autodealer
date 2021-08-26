package util

import (
	"testing"
)

func TestLocation(t *testing.T) {
	location := Location()
	x := "autodealer/util.TestLocation"
	if location != x {
		t.Errorf("Expected Location to be '%s'. Got '%s' instead\n", x, location)
	}
}

func TestLocation2(t *testing.T) {
	location := Location2()
	x := "testing.tRunner"
	if location != x {
		t.Errorf("Expected Location to be '%s'. Got '%s' instead\n", x, location)
	}
}