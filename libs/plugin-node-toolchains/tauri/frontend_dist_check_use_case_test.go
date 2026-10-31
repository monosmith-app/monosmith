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
	for dist, ok := range map[string]bool{"../../../dist/apps/studio": true, "../dist": false} {
		write(t, ws, "apps/studio/src-tauri/tauri.conf.json", `{"build":{"frontendDist":"`+dist+`"}}`)
		found := frontendDistFindings(ws, projects)
		if len(found) != 1 || found[0].OK != ok || !found[0].Warning || (!ok && !strings.Contains(found[0].Remedy, `"../../../dist/apps/studio"`)) {
			t.Errorf("%s: %+v", dist, found)
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
