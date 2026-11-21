package projectimport

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/outsidefiles"
)

func astarPlan() Plan {
	return Plan{Name: "astar", Folder: "packages/astar", Ecosystem: "dotnet", Source: Source{Path: "Roy-T.AStar"}, Outside: []outsidefiles.File{
		{Ref: `..\LICENSE`, Path: "LICENSE", Manifest: "Roy-T.AStar.csproj", Kind: outsidefiles.KindFile, Content: []byte("MIT")},
		{Path: "Directory.Build.props", Kind: outsidefiles.KindAbove},
	}}
}

func TestSettleOutsideTakesTheFlagAndRefusesAnUnknownOne(t *testing.T) {
	p := astarPlan()
	if err := SettleOutside(&p, Options{Outside: OutsideCopy, Yes: true}); err != nil || !p.CopyOutside {
		t.Fatalf("--outside-files copy: %v %+v", err, p.CopyOutside)
	}
	p = astarPlan()
	if err := SettleOutside(&p, Options{Yes: true}); err != nil || p.CopyOutside {
		t.Fatalf("--yes without the flag leaves the files: %v", err)
	}
	if err := SettleOutside(&p, Options{Outside: "move"}); err == nil || !strings.Contains(err.Error(), "--outside-files") {
		t.Fatalf("an unknown answer is refused: %v", err)
	}
}

func TestSummaryAndChangesNameTheOutsideFiles(t *testing.T) {
	p := astarPlan()
	p.CopyOutside = true
	text := Summary(p)
	for _, want := range []string{`outside  ..\LICENSE (Roy-T.AStar.csproj names it): copied into packages/astar as LICENSE`, "MSBuild reads Directory.Build.props from a folder above Roy-T.AStar"} {
		if !strings.Contains(text, want) {
			t.Errorf("the summary lacks %q:\n%s", want, text)
		}
	}
	found := false
	for _, g := range Changes(p).Groups {
		if g.Heading == "Outside files" && len(g.Lines) == 2 && g.Lines[0].Path == "packages/astar/LICENSE" && g.Lines[1].Path == "packages/astar/Roy-T.AStar.csproj" {
			found = true
		}
	}
	if !found {
		t.Errorf("the plan shows the copy and the changed project file: %+v", Changes(p).Groups)
	}
}

func TestBringOutsideCopiesOrNamesEachFile(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "packages", "astar")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "Roy-T.AStar.csproj"), []byte(`<None Include="..\LICENSE" />`), 0o644)

	p := astarPlan()
	left := BringOutside(root, p, &bytes.Buffer{})
	if len(left) != 2 || !strings.Contains(left[0].Problem, `the build needs LICENSE from the source (Roy-T.AStar.csproj names ..\LICENSE)`) || !strings.Contains(left[0].Remedy, "--outside-files copy") {
		t.Fatalf("leave: each file is named: %+v", left)
	}

	p.CopyOutside = true
	var out bytes.Buffer
	left = BringOutside(root, p, &out)
	if len(left) != 1 || !strings.Contains(left[0].Problem, "Directory.Build.props") {
		t.Fatalf("copy: only the file above is left: %+v", left)
	}
	if raw, _ := os.ReadFile(filepath.Join(dir, "LICENSE")); string(raw) != "MIT" {
		t.Fatalf("LICENSE: %q", raw)
	}
	if raw, _ := os.ReadFile(filepath.Join(dir, "Roy-T.AStar.csproj")); string(raw) != `<None Include="LICENSE" />` {
		t.Fatalf("the project file: %s", raw)
	}
	if !strings.Contains(out.String(), "copied LICENSE into packages/astar") {
		t.Fatalf("progress: %s", out.String())
	}
}
