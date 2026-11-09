package commands

import "testing"

func TestKeyTakenSeesGlobalAndViewKeysOnly(t *testing.T) {
	for _, tc := range []struct {
		key, view string
		want      bool
	}{
		{"b", TabProjects, true}, // global
		{"e", TabProjects, true}, // Projects: explain
		{"e", TabRelease, false}, // only Doctor and Projects use it
		{"s", TabProjects, false},
		{"s", TabDeps, false}, // Deps no longer has the safe key
		{"p", TabDeps, true},  // Deps: the Patch chip
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
				if exclusive(a, b) {
					continue // one is in use only where the other is not
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

// exclusive is true for two commands that are never in use together: one needs
// the run panel (or the recent runs) and the other leaves while it is there.
func exclusive(a, b Command) bool {
	return (a.NeedsRunPanel && b.NotInRunPanel) || (b.NeedsRunPanel && a.NotInRunPanel) ||
		(a.NeedsRecentFocus && b.NotInRecentFocus) || (b.NeedsRecentFocus && a.NotInRecentFocus) ||
		(a.NeedsSideFocus && b.NotInSideFocus) || (b.NeedsSideFocus && a.NotInSideFocus) ||
		(a.NeedsSideNext && b.NotSideNext) || (b.NeedsSideNext && a.NotSideNext)
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
