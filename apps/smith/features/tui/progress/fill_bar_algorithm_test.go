package progress

import "testing"

func TestFilled(t *testing.T) {
	cases := []struct{ done, total, width, want int }{
		{0, 10, 20, 0}, {5, 10, 20, 10}, {10, 10, 20, 20}, {1, 100, 20, 1}, {99, 100, 20, 19},
		{3, 0, 20, 0}, {12, 10, 20, 20}, {5, 10, 0, 0}, {-1, 10, 20, 0},
	}
	for _, c := range cases {
		if got := Filled(c.done, c.total, c.width); got != c.want {
			t.Errorf("Filled(%d,%d,%d) = %d, want %d", c.done, c.total, c.width, got, c.want)
		}
	}
}

func TestPercent(t *testing.T) {
	if Percent(1, 3) != 33 || Percent(3, 3) != 100 || Percent(0, 0) != 0 || Percent(9, 3) != 100 {
		t.Fatal("percent")
	}
}
