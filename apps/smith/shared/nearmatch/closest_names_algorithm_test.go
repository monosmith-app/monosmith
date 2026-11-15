package nearmatch

import (
	"reflect"
	"testing"
)

func TestClosest(t *testing.T) {
	names := []string{"web", "api", "core", "webhooks"}
	for _, c := range []struct {
		in   string
		want []string
	}{
		{"wbe", []string{"web"}},
		{"Web", []string{"web", "webhooks"}},
		{"zzzzzz", []string{}},
	} {
		if got := Closest(c.in, names); !reflect.DeepEqual(got, c.want) {
			t.Errorf("Closest(%q) = %v, want %v", c.in, got, c.want)
		}
	}
	if Suggestion("api2", names) != " Did you mean api?" || Suggestion("qqqqqq", names) != "" {
		t.Error("suggestion sentence")
	}
}
