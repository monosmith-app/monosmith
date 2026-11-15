package clicontract

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceReport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `[{"tag_name": "packages/plugin-sdk/v0.9.0"}, {"tag_name": "smith@2.0.0-rc.1", "prerelease": true}, {"tag_name": "smith@1.4.0"}, {"tag_name": "smith@1.10.0"}, {"tag_name": "smith@1.9.0"}]`)
	}))
	defer srv.Close()
	releasesURL = srv.URL
	t.Setenv("SMITH_NO_UPDATE_CHECK", "")
	// Under smith (CI runs the tests as a task) this names the outer workspace.
	t.Setenv("SMITH_WORKSPACE_ROOT", "")

	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "smith.toml"), []byte("[workspace]\nname = \"acme\"\n"), 0o644)
	info := WorkspaceReport("1.2.0", ws)
	if info.CLI.Latest == nil || *info.CLI.Latest != "1.10.0" || !info.CLI.UpdateAvailable {
		t.Fatalf("cli %+v", info.CLI)
	}
	if info.Workspace == nil || info.Workspace.Config["workspace"].(map[string]any)["name"] != "acme" {
		t.Fatalf("workspace %+v", info.Workspace)
	}
	if info := WorkspaceReport("dev", t.TempDir()); info.CLI.UpdateAvailable || info.Workspace != nil {
		t.Fatalf("a development build outside a workspace: %+v", info)
	}
	t.Setenv("SMITH_NO_UPDATE_CHECK", "1")
	if info := WorkspaceReport("1.2.0", ws); info.CLI.Latest != nil {
		t.Fatalf("the check can be turned off: %+v", info.CLI)
	}
}
