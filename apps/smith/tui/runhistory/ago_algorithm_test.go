package runhistory

import (
	"testing"
	"time"
)

func TestAgoNamesTheAgeInTheLargestFittingUnit(t *testing.T) {
	for _, c := range []struct {
		d    time.Duration
		want string
	}{
		{-time.Second, "just now"},
		{4 * time.Second, "just now"},
		{40 * time.Second, "40s ago"},
		{3 * time.Minute, "3 min ago"},
		{59*time.Minute + 59*time.Second, "59 min ago"},
		{2 * time.Hour, "2 h ago"},
		{49 * time.Hour, "2 d ago"},
	} {
		if got := Ago(c.d); got != c.want {
			t.Errorf("Ago(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}
