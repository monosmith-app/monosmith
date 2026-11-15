package dependencies

import (
	"path/filepath"
	"testing"
)

func TestDefaultTargetIsTheProjectHereElseTheWorkspaceRoot(t *testing.T) {
	ws := filepath.FromSlash("/repo")
	roots := map[string]string{"web": "apps/web", "webkit": "apps/webkit", "docs": "apps/web/docs"}
	for cwd, want := range map[string]string{
		"/repo":                   RootTarget,
		"/repo/tools":             RootTarget,
		"/repo/apps/web":          "web",
		"/repo/apps/web/src/ui":   "web",
		"/repo/apps/webkit":       "webkit",
		"/repo/apps/web/docs/a":   "docs",
		"/elsewhere/apps/web/src": RootTarget,
	} {
		if got := DefaultTarget(ws, filepath.FromSlash(cwd), roots); got != want {
			t.Errorf("%s: got %q, want %q", cwd, got, want)
		}
	}
}
