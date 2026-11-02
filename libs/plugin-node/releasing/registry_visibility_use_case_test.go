package releasing

import (
	"reflect"
	"testing"
	"time"
)

func TestSameReleaseOptionalDependencies(t *testing.T) {
	manifest := []byte(`{"name":"@acme/cli","optionalDependencies":{"@acme/cli-linux":"1.2.0","@acme/cli-win":"1.2.0","left-pad":"^1.0.0","@acme/old":"1.1.0"}}`)
	if got := SameReleaseOptionalDependencies(manifest, "1.2.0"); !reflect.DeepEqual(got, []string{"@acme/cli-linux", "@acme/cli-win"}) {
		t.Errorf("%v", got)
	}
	if got := SameReleaseOptionalDependencies([]byte(`not json`), "1.2.0"); got != nil {
		t.Errorf("%v", got)
	}
}

// fakeRegistry shows a package after it was asked about n times; the clock
// moves only when Sleep is called.
type fakeRegistry struct {
	asks    map[string]int
	showAt  map[string]int
	elapsed time.Duration
}

func (f *fakeRegistry) visible(name, _ string) bool {
	f.asks[name]++
	at, ok := f.showAt[name]
	return ok && f.asks[name] >= at
}

func TestAwaitWaitsForPackagesThatAppearLate(t *testing.T) {
	f := &fakeRegistry{asks: map[string]int{}, showAt: map[string]int{"a": 1, "b": 3}}
	v := Visibility{Visible: f.visible, Wait: 60 * time.Second, Every: 5 * time.Second, Sleep: func(d time.Duration) { f.elapsed += d }}
	if missing := v.Await([]string{"a", "b"}, "1.0.0"); len(missing) != 0 {
		t.Fatalf("still missing %v", missing)
	}
	if f.elapsed != 10*time.Second || f.asks["a"] != 1 {
		t.Errorf("waited %v, asked a %d times (a is visible at once and is not asked again)", f.elapsed, f.asks["a"])
	}
}

func TestAwaitGivesUpAfterTheWaitAndNamesWhatIsMissing(t *testing.T) {
	f := &fakeRegistry{asks: map[string]int{}, showAt: map[string]int{"a": 1}}
	v := Visibility{Visible: f.visible, Wait: 60 * time.Second, Every: 10 * time.Second, Sleep: func(d time.Duration) { f.elapsed += d }}
	missing := v.Await([]string{"a", "gone"}, "1.0.0")
	if !reflect.DeepEqual(missing, []string{"gone"}) || f.elapsed != 60*time.Second {
		t.Errorf("missing %v after %v", missing, f.elapsed)
	}
}
