package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestDefaultsVintageFinding(t *testing.T) {
	latest := workspace.LatestDefaultsVintage()
	if f := defaultsVintageFinding(latest); !f.OK {
		t.Errorf("latest is not ok: %+v", f)
	}
	if f := defaultsVintageFinding(""); f.OK || !f.Warning || !strings.Contains(f.Remedy, "smith upgrade --defaults") {
		t.Errorf("none recorded: %+v", f)
	}
	if f := defaultsVintageFinding("2999.01"); f.OK || f.Remedy != "upgrade smith" {
		t.Errorf("newer: %+v", f)
	}
}
