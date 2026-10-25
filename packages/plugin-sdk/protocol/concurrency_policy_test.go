package protocol

import "testing"

func TestConcurrencyDefaultsToOneAtATime(t *testing.T) {
	for n, want := range map[int]int{0: 1, 1: 1, 4: 4, -2: 1, -100: 1} {
		if limit, unlimited := ConcurrencyLimit(n); limit != want || unlimited {
			t.Errorf("ConcurrencyLimit(%d) = %d, %v; want %d, false", n, limit, unlimited, want)
		}
	}
}

func TestMinusOneLeavesTheNumberToTheHost(t *testing.T) {
	if _, unlimited := ConcurrencyLimit(ConcurrencyUnlimited); !unlimited {
		t.Fatal("-1 must be unlimited")
	}
}

func TestOnlyDefinedConcurrenciesAreValid(t *testing.T) {
	for n, want := range map[int]bool{-2: false, -1: true, 0: true, 1: true, 64: true} {
		if got := ConcurrencyValid(n); got != want {
			t.Errorf("ConcurrencyValid(%d) = %v, want %v", n, got, want)
		}
	}
}
