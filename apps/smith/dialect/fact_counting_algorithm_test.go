package dialect

import (
	"slices"
	"testing"
)

func shapes(key string, values ...string) []map[string]string {
	var out []map[string]string
	for _, v := range values {
		out = append(out, map[string]string{key: v})
	}
	return out
}

func TestADominantValueNeedsThreeSiblingsAndTwoThirds(t *testing.T) {
	cases := []struct {
		name   string
		values []string
		want   string
		ok     bool
		split  bool
	}{
		{"all agree", []string{"a", "a", "a"}, "a", true, false},
		{"two thirds of six", []string{"a", "a", "a", "a", "b", "c"}, "a", true, false},
		{"just under two thirds", []string{"a", "a", "a", "b", "b"}, "a", false, true},
		{"only two siblings", []string{"a", "a"}, "a", false, false},
		{"a tie", []string{"a", "a", "a", "b", "b", "b"}, "a", false, true},
		{"one outlier of twenty-two", slices.Concat([]string{"x"}, slices.Repeat([]string{"a"}, 21)), "a", true, false},
	}
	for _, c := range cases {
		facts := CountFacts(shapes("k", c.values...))
		if len(facts) != 1 {
			t.Fatalf("%s: facts %v", c.name, facts)
		}
		got, ok := facts[0].Dominant()
		if ok != c.ok || (ok && got.Value != c.want) || facts[0].Split() != c.split {
			t.Errorf("%s: dominant %v %v, split %v", c.name, got, ok, facts[0].Split())
		}
	}
}

func TestCountingOnlyCountsTheSiblingsThatReportAKey(t *testing.T) {
	in := []map[string]string{{"a": "1", "b": "x"}, {"a": "1"}, {"a": "1"}, nil}
	facts := CountFacts(in)
	if len(facts) != 2 || facts[0].Key != "a" || facts[0].Total != 3 || facts[1].Key != "b" || facts[1].Total != 1 {
		t.Fatalf("facts %+v", facts)
	}
	if facts[1].Split() {
		t.Error("one sibling is no habit")
	}
}
