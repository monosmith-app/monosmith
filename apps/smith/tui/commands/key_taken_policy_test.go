package commands

import "testing"

func TestKeyTakenSeesGlobalAndViewKeysOnly(t *testing.T) {
	for _, tc := range []struct {
		key, view string
		want      bool
	}{
		{"b", TabProjects, true}, // global
		{"e", TabProjects, true}, // Projects: explain
		{"e", TabRun, false},     // only Doctor and Projects use it
		{"s", TabProjects, false},
		{"s", TabDeps, true},
	} {
		if got := KeyTaken(tc.key, tc.view); got != tc.want {
			t.Errorf("KeyTaken(%q, %s) = %v, want %v", tc.key, tc.view, got, tc.want)
		}
	}
}

func TestNoViewGivesTwoCommandsTheSameKey(t *testing.T) {
	for i, a := range all {
		for _, b := range all[i+1:] {
			for _, k := range a.Keys {
				if !contains(b.Keys, k) || a.Kind == PaneKey || b.Kind == PaneKey {
					continue
				}
				if len(a.Views) == 0 || len(b.Views) == 0 {
					continue // a view's own key wins over the global one by design
				}
				for _, v := range a.Views {
					if contains(b.Views, v) {
						t.Errorf("%s and %s both use %q in %s", a.ID, b.ID, k, v)
					}
				}
			}
		}
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
