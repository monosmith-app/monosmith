package buildoutputs

import "testing"

func TestInsideAnyMatchesThePathAndWhatIsUnderIt(t *testing.T) {
	allowed := []string{"apps/mvd/appwindow/web/"}
	for p, want := range map[string]bool{
		"apps/mvd/appwindow/web":      true,
		"apps/mvd/appwindow/web/a.js": true,
		"apps/mvd/appwindow/webextra": false,
		"apps/mvd/appwindow":          false,
	} {
		if got := InsideAny(p, allowed); got != want {
			t.Errorf("%s: got %v, want %v", p, got, want)
		}
	}
}
