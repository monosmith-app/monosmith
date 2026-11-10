package dependencies

import "testing"

func TestTheSummaryCountsTheUpdatesAndSplitsThemByLevel(t *testing.T) {
	cases := []struct {
		name    string
		updates []Update
		want    string
	}{
		{"none", nil, ""},
		{"one", []Update{{Level: "patch"}}, "1 update: 1 patch"},
		{"mixed", []Update{{Level: "minor"}, {Level: "patch"}, {Level: "minor"}}, "3 updates: 1 patch, 2 minor"},
		{"all three levels", []Update{{Level: "major"}, {Level: "patch"}, {Level: "minor"}}, "3 updates: 1 patch, 1 minor, 1 major"},
		{"an unknown level comes last", []Update{{Level: "odd"}, {Level: "patch"}}, "2 updates: 1 patch, 1 odd"},
	}
	for _, c := range cases {
		if got := UpdateSummary(c.updates); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}
