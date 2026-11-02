package taskrun

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
)

func TestASharedOutputFolderKeepsTheFilesAnotherTaskWrote(t *testing.T) {
	for name, tc := range map[string]struct {
		shared   []string
		wantKept bool
	}{
		"listed in shared_outputs": {[]string{"apps/web/out"}, true},
		"not listed":               {nil, false},
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t, "core")
			keep := filepath.Join(f.root, "apps", "web", "out", "earlier.zip")
			_ = os.MkdirAll(filepath.Dir(keep), 0o755)
			if err := os.WriteFile(keep, []byte("zip"), 0o644); err != nil {
				t.Fatal(err)
			}
			g, err := taskgraph.Build(f.pg, []string{"web"}, []string{"build"})
			if err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			s := Run(context.Background(), g, Options{Root: f.root, Files: f.files, Projects: f.pg, Storage: f.store, Out: &out, ToolVersion: "test", SharedOutputs: tc.shared})
			if !s.OK() {
				t.Fatalf("run failed:\n%s", out.String())
			}
			_, err = os.Stat(keep)
			if kept := err == nil; kept != tc.wantKept {
				t.Errorf("earlier file kept = %v, want %v", kept, tc.wantKept)
			}
		})
	}
}
