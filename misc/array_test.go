package misc

import (
	"testing"
)

func TestStringInSlice(t *testing.T) {
	if StringInSlice("a", []string{"c", "b", "a"}) == false {
		t.Errorf("Expected true error, got false")
	}
	if StringInSlice("d", []string{"c", "b", "a"}) {
		t.Errorf("Expected false error, got true")
	}
}
