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
	if got["Node dependencies are installed"] || !got["npm workspaces include apps/web"] || got["npm workspaces include tools/x"] {
		t.Fatalf("findings %+v", res.Findings)
	}
	for _, f := range res.Findings {
		if f.Check == "npm workspaces include tools/x" && !strings.Contains(f.Remedy, `exclude = ["tools/x"]`) {
			t.Errorf("the remedy does not offer the exclusion: %s", f.Remedy)
		}
	}
}
