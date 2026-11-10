package npmpropagation

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time        { return c.now }
func (c *fakeClock) Sleep(d time.Duration) { c.now = c.now.Add(d) }

// fakeRegistry answers a package only after its visible-from time.
type fakeRegistry struct {
	clock     *fakeClock
	start     time.Time
	from      map[string]time.Duration
	noTarball map[string]bool
	probes    int
}

func (r *fakeRegistry) Lookup(name, _ string) (bool, string) {
	r.probes++
	from, ok := r.from[name]
	if !ok || r.clock.now.Sub(r.start) < from {
		return false, ""
	}
	return true, "https://registry/" + name + ".tgz"
}

func (r *fakeRegistry) Fetchable(url string) bool {
	return !r.noTarball[strings.TrimSuffix(strings.TrimPrefix(url, "https://registry/"), ".tgz")]
}

func setup(from map[string]time.Duration) (*fakeClock, *fakeRegistry) {
	clock := &fakeClock{now: time.Unix(1000, 0)}
	return clock, &fakeRegistry{clock: clock, start: clock.now, from: from}
}

func TestAwaitNeedsThreeConfirmationsSpacedApart(t *testing.T) {
	clock, reg := setup(map[string]time.Duration{"a": 0, "b": 0})
	out := Await(reg, clock, Config{Max: 15 * time.Minute}, []string{"a", "b"}, "1.0.0")
	if !out.Ready || out.Waited != 60*time.Second {
		t.Fatalf("outcome %+v: want ready after two 30 s pauses", out)
	}
}

func TestAwaitWaitsForALatePackageThenConfirms(t *testing.T) {
	clock, reg := setup(map[string]time.Duration{"a": 0, "b": 5 * time.Minute})
	var lines []string
	out := Await(reg, clock, Config{Max: 15 * time.Minute, Progress: func(l string) { lines = append(lines, l) }}, []string{"a", "b"}, "1.0.0")
	if !out.Ready || out.Waited < 5*time.Minute+60*time.Second {
		t.Fatalf("outcome %+v", out)
	}
	if len(lines) < 4 || !strings.Contains(lines[0], "not fetchable yet: b") {
		t.Fatalf("progress lines %q: want one a minute naming b", lines)
	}
}

func TestAwaitGivesUpAtTheMaximumAndNamesWhatIsMissing(t *testing.T) {
	clock, reg := setup(map[string]time.Duration{"a": 0, "b": time.Hour})
	out := Await(reg, clock, Config{Max: 2 * time.Minute}, []string{"a", "b"}, "1.0.0")
	if out.Ready || len(out.Missing) != 1 || out.Missing[0] != "b" || out.Waited != 2*time.Minute {
		t.Fatalf("outcome %+v", out)
	}
}

func TestAwaitStopsWhenNotConfirmedInTime(t *testing.T) {
	clock, reg := setup(map[string]time.Duration{"a": 0})
	out := Await(reg, clock, Config{Max: 45 * time.Second}, []string{"a"}, "1.0.0")
	if out.Ready || len(out.Missing) != 0 {
		t.Fatalf("outcome %+v: visible but not confirmed three times", out)
	}
}

func TestAMissingTarballCountsAsMissing(t *testing.T) {
	clock, reg := setup(map[string]time.Duration{"a": 0})
	reg.noTarball = map[string]bool{"a": true}
	out := Await(reg, clock, Config{Max: time.Minute}, []string{"a"}, "1.0.0")
	if out.Ready || len(out.Missing) != 1 {
		t.Fatalf("outcome %+v", out)
	}
}

type flakyRegistry struct {
	*fakeRegistry
	failOn, calls int
}

func (f *flakyRegistry) Lookup(n, v string) (bool, string) {
	f.calls++
	if f.calls == f.failOn {
		return false, ""
	}
	return f.fakeRegistry.Lookup(n, v)
}

func TestAFailedProbeRestartsTheCount(t *testing.T) {
	clock, reg := setup(map[string]time.Duration{"a": 0})
	flaky := &flakyRegistry{fakeRegistry: reg, failOn: 2}
	out := Await(flaky, clock, Config{Max: 15 * time.Minute}, []string{"a"}, "1.0.0")
	if !out.Ready || flaky.calls < 5 {
		t.Fatalf("outcome %+v after %d probes: a failure in the middle needs three new passes", out, flaky.calls)
	}
}

func TestMaxZeroProbesOnceAndDoesNotSleep(t *testing.T) {
	clock, reg := setup(map[string]time.Duration{"a": 0, "b": time.Hour})
	out := Await(reg, clock, Config{}, []string{"a", "b"}, "1.0.0")
	if out.Ready || out.Waited != 0 || reg.probes != 2 {
		t.Fatalf("outcome %+v probes %d", out, reg.probes)
	}
	clock, reg = setup(map[string]time.Duration{"a": 0})
	if out := Await(reg, clock, Config{}, []string{"a"}, "1.0.0"); !out.Ready || out.Waited != 0 {
		t.Fatalf("outcome %+v", out)
	}
}

func TestParseWait(t *testing.T) {
	for in, want := range map[string]time.Duration{"": DefaultMax, "0": 0, "90s": 90 * time.Second, "15m": 15 * time.Minute} {
		if got, err := ParseWait(in); err != nil || got != want {
			t.Errorf("%q: %v %v", in, got, err)
		}
	}
	for _, in := range []string{"soon", "-5m", "10"} {
		if _, err := ParseWait(in); err == nil {
			t.Errorf("%q must be refused", in)
		}
	}
}

func TestNPMLookupAndFetchable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead || r.URL.Path != "/a.tgz" {
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	n := NPM{Output: func(_ string, argv []string) (string, error) {
		if !strings.Contains(strings.Join(argv, " "), "--prefer-online") {
			t.Errorf("argv %v: must bypass the cache", argv)
		}
		return `{"version":"1.0.0","dist.tarball":"` + srv.URL + `/a.tgz"}`, nil
	}}
	found, url := n.Lookup("a", "1.0.0")
	if !found || !n.Fetchable(url) || n.Fetchable(srv.URL+"/b.tgz") {
		t.Fatalf("found %v url %q", found, url)
	}
	n.Output = func(string, []string) (string, error) { return "", nil }
	if found, _ := n.Lookup("a", "1.0.0"); found {
		t.Fatal("an empty answer is a missing version")
	}
}
