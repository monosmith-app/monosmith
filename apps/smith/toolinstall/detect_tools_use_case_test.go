package toolinstall

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// fakeDetector answers from a table and never runs a program.
type fakeDetector map[string]Probe

func (f fakeDetector) Detect(_ context.Context, n Need) Probe { return f[n.Name] }

// hangingDetector blocks until the context ends, like a program that hangs.
type hangingDetector struct{}

func (hangingDetector) Detect(ctx context.Context, _ Need) Probe {
	<-ctx.Done()
	return Probe{}
}

func TestDetectAllJudgesVersionsAgainstThePin(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module m\n\ngo 1.23.4\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	goTool := Need{ToolNeed: protocol.ToolNeed{Name: "go", Pins: []protocol.ToolVersionPin{{File: "go.mod", Format: "go-mod"}}}, Dirs: []string{root}}
	py := Need{ToolNeed: protocol.ToolNeed{Name: "python"}, Dirs: []string{root}}
	node := Need{ToolNeed: protocol.ToolNeed{Name: "node"}, Dirs: []string{root}}
	d := fakeDetector{"go": {Program: "go", Version: "1.22.0"}, "python": {Program: "python3", Version: "3.13.1"}}
	got := DetectAll(context.Background(), d, []Need{goTool, py, node})
	want := []State{Outdated, Found, Missing}
	for i, w := range want {
		if got[i].State != w {
			t.Errorf("%s: state %s, want %s", got[i].Name, got[i].State, w)
		}
	}
	if got[0].WantedFrom != "go.mod" || got[0].Wanted.Min != "1.23.4" {
		t.Errorf("go wanted %+v from %q", got[0].Wanted, got[0].WantedFrom)
	}
}

func TestDetectAllNeverWaitsLongerThanTheLimit(t *testing.T) {
	start := time.Now()
	got := DetectAll(context.Background(), hangingDetector{}, []Need{{ToolNeed: protocol.ToolNeed{Name: "a"}}, {ToolNeed: protocol.ToolNeed{Name: "b"}}})
	if elapsed := time.Since(start); elapsed > DetectTimeout+time.Second {
		t.Fatalf("detection took %s", elapsed)
	}
	if got[0].State != Missing || got[1].State != Missing {
		t.Fatalf("a hanging program must count as missing: %+v", got)
	}
}
