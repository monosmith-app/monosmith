package adoption

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func inspectWith(t *testing.T, files ...string) Report {
	t.Helper()
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Skipf("git unavailable: %v %s", err, out)
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r, err := Inspect(dir, "0.0.0", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestAdoptSendsNxAndTurborepoToConvert(t *testing.T) {
	for _, c := range []struct{ file, say, step string }{
		{"turbo.json", "Turborepo", "smith convert --dry-run"},
		{"nx.json", "Nx", "npm ci"},
	} {
		r := inspectWith(t, c.file)
		if len(r.Blockers) != 1 || !strings.Contains(r.Blockers[0].Message, c.say) || !strings.Contains(strings.Join(append(r.Blockers[0].Steps, r.Blockers[0].Step), " "), c.step) || r.Adoptable() {
			t.Errorf("%s: %+v", c.file, r.Blockers)
		}
	}
}
