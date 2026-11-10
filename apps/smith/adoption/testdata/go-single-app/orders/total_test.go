package orders

import "testing"

func TestTotal(t *testing.T) {
	if got := Total([]float64{2, 3}); got != 5 {
		t.Fatalf("Total = %v, want 5", got)
	}
}
