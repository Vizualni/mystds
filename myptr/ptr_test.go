package myptr

import "testing"

// sigh...
func TestRefForReal(t *testing.T) {
	ptr := Ref(42)
	if *ptr != 42 {
		t.Errorf("Expected %d, got %d", 42, *ptr)
	}
}
