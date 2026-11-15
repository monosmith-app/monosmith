package scaffold

import "testing"

func TestSameChoicesIgnoresOrderAndNeverMatchesNothing(t *testing.T) {
	for _, tc := range []struct {
		a, b []string
		want bool
	}{
		{[]string{"react", "vitest"}, []string{"vitest", "react"}, true},
		{[]string{"react"}, []string{"react", "vitest"}, false},
		{[]string{"react", "vitest"}, []string{"vue", "vitest"}, false},
		{nil, nil, false},
	} {
		if got := sameChoices(tc.a, tc.b); got != tc.want {
			t.Errorf("sameChoices(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
