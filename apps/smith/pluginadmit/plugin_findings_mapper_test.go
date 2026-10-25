package pluginadmit

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
)

func TestFindings(t *testing.T) {
	ok := Status{Plugin: Plugin{Name: "kiwi", Source: ".smith/plugins"}, Env: []string{"AWS_*"}, Trusted: true}
	untrusted := Status{Plugin: Plugin{Name: "widget", Source: "smith.toml"}, Pinned: true, Reason: plugintrust.ReasonNew}

	got := Findings([]Status{ok, untrusted}, false)
	if len(got) != 2 {
		t.Fatalf("got %d findings", len(got))
	}
	if !got[0].OK || got[0].Check != "external plugin kiwi is trusted (from .smith/plugins, pinned no, sees the baseline plus AWS_*)" {
		t.Errorf("passing finding: %+v", got[0])
	}
	f := got[1]
	if f.OK || f.Check != "external plugin widget is trusted" {
		t.Fatalf("failing finding: %+v", f)
	}
	if !strings.Contains(f.Detail, "pinned yes") || !strings.Contains(f.Remedy, "smith plugin trust widget") {
		t.Errorf("detail %q remedy %q", f.Detail, f.Remedy)
	}

	ci := Findings([]Status{{Plugin: untrusted.Plugin, Reason: plugintrust.ReasonNotPinned}}, true)
	if !strings.Contains(ci[0].Remedy, "smith plugin trust widget --pin") {
		t.Errorf("CI remedy %q", ci[0].Remedy)
	}
}
