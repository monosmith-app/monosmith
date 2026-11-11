package tauri

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDoctorChecksFrontendDist(t *testing.T) {
	ws := t.TempDir()
	projects := []protocol.Project{{Name: "studio", Root: "apps/studio", Tasks: map[string]protocol.Task{"build": {}}}}
	// The shipped folder dist/apps/studio holds the installers, so the web app
	// is a work file in dist/tauri/apps/studio/frontend (ADR 0092).
	for dist, ok := range map[string]bool{"../../../dist/tauri/apps/studio/frontend": true, "../../../dist/apps/studio": false, "../dist": false} {
		write(t, ws, "apps/studio/src-tauri/tauri.conf.json", `{"build":{"frontendDist":"`+dist+`"}}`)
		found := frontendDistFindings(ws, projects)
		if len(found) != 1 || found[0].OK != ok || !found[0].Warning || (!ok && !strings.Contains(found[0].Remedy, `"../../../dist/tauri/apps/studio/frontend"`)) {
			t.Errorf("%s: %+v", dist, found)
		}
		// Only the shipped folder makes the web build uncached until the move.
		if uncached := strings.Contains(found[0].Remedy, "does not cache the web build"); uncached != (dist == "../../../dist/apps/studio") {
			t.Errorf("%s: remedy %q", dist, found[0].Remedy)
		}
	}
	write(t, ws, "apps/studio/src-tauri/tauri.conf.json", `{"build":{"frontendDist":"http://localhost:5173"}}`)
	if found := frontendDistFindings(ws, projects); len(found) != 0 {
		t.Errorf("a URL is not checked: %+v", found)
	}
}

func write(t *testing.T, ws, rel, body string) {
	t.Helper()
	p := filepath.Join(ws, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
