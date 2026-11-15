package cachemaintenance

import (
	"testing"
	"time"
)

func TestShouldCheck(t *testing.T) {
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	const limit = 100 << 20
	measured := func(ago time.Duration, size, added int64) pruneState {
		return pruneState{LastCheck: now.Add(-ago), SizeBytes: size, AddedBytes: added}
	}
	for _, tc := range []struct {
		name  string
		state pruneState
		added int64
		limit int64
		want  bool
	}{
		{"no limit never checks", pruneState{}, 50 << 20, 0, false},
		{"a negative limit is no limit", pruneState{}, 50 << 20, -1, false},
		{"a store never measured is measured", pruneState{}, 1, limit, true},
		{"measured an hour ago is measured again", measured(time.Hour, 1, 0), 1, limit, true},
		{"measured a minute ago and a small run is left alone", measured(time.Minute, 99<<20, 0), 1 << 20, limit, false},
		{"a big run over the limit is checked at once", measured(time.Minute, 99<<20, 0), 5 << 20, limit, true},
		{"a big run that still fits is left alone", measured(time.Minute, 10<<20, 0), 5 << 20, limit, false},
		{"earlier runs count toward the total", measured(time.Minute, 90<<20, 8<<20), 5 << 20, limit, true},
		{"a clock set back is not an hour", measured(-time.Hour, 1, 0), 1, limit, false},
	} {
		if got := shouldCheck(tc.state, now, tc.added, tc.limit); got != tc.want {
			t.Errorf("%s: shouldCheck = %v, want %v", tc.name, got, tc.want)
		}
	}
}
