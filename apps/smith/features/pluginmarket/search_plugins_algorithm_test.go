package pluginmarket

import "testing"

func names(ps []Plugin) []string {
	var out []string
	for _, p := range ps {
		out = append(out, p.Name)
	}
	return out
}

func TestSearchMatchesEveryWordInAnyField(t *testing.T) {
	idx := mustIndex(t, sampleIndexJSON)
	cases := map[string][]string{
		"":              {"deployer", "widget"},
		"WIDGET":        {"widget"},
		"moon":          {"deployer"},
		"deploy":        {"deployer"},
		"scaffold go":   {"widget"},
		"scaffold moon": nil,
		"nothing":       nil,
	}
	for term, want := range cases {
		got := names(Search(idx, term))
		if len(got) != len(want) {
			t.Fatalf("%q: got %v, want %v", term, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%q: got %v, want %v", term, got, want)
			}
		}
	}
}
