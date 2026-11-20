package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDoctor(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "package.json"), []byte(`{"workspaces":["apps/*"]}`), 0o644)
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{{Root: "apps/web"}, {Root: "tools/x"}}}, "24")
	got := map[string]bool{}
	for _, f := range res.Findings {
		got[f.Check] = f.OK
	}
	if ok, listed := got["Node dependencies are installed"]; ok || !listed {
		t.Fatalf("findings %+v", res.Findings)
	}
	for check := range got {
		if strings.HasPrefix(check, "npm workspaces include ") {
			t.Errorf("the workspaces coverage is the host's join check now, not a plugin finding: %s", check)
		}
	}
}
