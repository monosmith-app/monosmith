package release

import (
	"reflect"
	"testing"
)

func TestPlannedItems(t *testing.T) {
	plan := Plan{Items: []Item{{Project: "core", Next: "1.5.0"}, {Project: "api", Next: "6.2.0"}, {Project: "app", Next: "6.2.0-beta.1"}}}
	cases := []struct {
		name     string
		projects []string
		want     []string
	}{
		{"none packed", nil, nil},
		{"only planned projects", []string{"api", "web"}, []string{"api"}},
		{"plan order kept", []string{"app", "core"}, []string{"core", "app"}},
		{"not in the plan", []string{"web"}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got []string
			for _, item := range PlannedItems(plan, c.projects) {
				got = append(got, item.Project)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("PlannedItems(%v) = %v, want %v", c.projects, got, c.want)
			}
		})
	}
}
