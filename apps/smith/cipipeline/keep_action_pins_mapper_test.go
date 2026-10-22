package cipipeline

import "testing"

func TestKeepNewerPinsKeepsWhatABotBumpedAndNothingElse(t *testing.T) {
	old := "      - uses: actions/cache@v6\n      - uses: actions/checkout@v7\n      - uses: actions/setup-go@v5\n      - uses: dart-lang/setup-dart@main\n      - uses: actions/cache@v6\n"
	rendered := "      - uses: actions/cache@v4\n      - uses: actions/checkout@v7\n      - uses: actions/setup-go@v7\n      - uses: dart-lang/setup-dart@v1\n      - uses: actions/new@v1\n"
	want := "      - uses: actions/cache@v6\n      - uses: actions/checkout@v7\n      - uses: actions/setup-go@v7\n      - uses: dart-lang/setup-dart@v1\n      - uses: actions/new@v1\n"
	if got := KeepNewerPins(old, rendered); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestNewerVersionComparesNumbersNotText(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{{"v10", "v9", true}, {"v9", "v10", false}, {"v4", "v4", false}, {"v4.1", "v4", true}, {"v4", "v4.0.0", false}, {"v1.2.3", "v1.10.0", false}}
	for _, c := range cases {
		if got := newerVersion(c.a, c.b); got != c.want {
			t.Errorf("newerVersion(%s, %s) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestPreserveKeepsSlotsAndPins(t *testing.T) {
	old := "      - uses: actions/cache@v6\n      # smith:slot before-setup\n      - run: echo mine\n      # smith:slot-end before-setup\n"
	rendered := "      - uses: actions/cache@v4\n      # smith:slot before-setup\n      # smith:slot-end before-setup\n"
	got, lost := Preserve(old, rendered)
	if got != old || len(lost) != 0 {
		t.Errorf("got %q, lost %v", got, lost)
	}
}
