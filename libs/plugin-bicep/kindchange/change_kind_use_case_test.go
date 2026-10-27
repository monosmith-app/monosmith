package kindchange

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func edited(r protocol.ChangeKindResult) map[string]string {
	out := map[string]string{}
	for _, f := range r.Files {
		out[f.Path] = f.Content
	}
	return out
}

const storageBicep = "metadata version = '0.0.0'\nmetadata smithKind = 'lib'\n\nmodule log '../util/main.bicep' = {\n  name: 'log'\n}\n"

var workspace = map[string]string{
	"libs/storage/main.bicep":    storageBicep,
	"libs/util/main.bicep":       "metadata smithKind = 'lib'\n",
	"apps/infra/main.bicep":      "metadata smithKind = 'app'\n\nmodule shared '../../libs/storage/main.bicep' = {\n  name: 'shared'\n}\n",
	"apps/infra/main.bicepparam": "using './main.bicep'\n",
}

func projects() []protocol.Project {
	return []protocol.Project{
		{Name: "storage", Root: "libs/storage", Language: "bicep"},
		{Name: "util", Root: "libs/util", Language: "bicep"},
		{Name: "infra", Root: "apps/infra", Language: "bicep"},
	}
}

func TestLibMovesToApps(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{Project: projects()[0], To: protocol.ProjectApp, NewRoot: "apps/storage", Projects: projects()})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)
	storage := got["libs/storage/main.bicep"]
	if !strings.Contains(storage, "metadata smithKind = 'app'") {
		t.Fatalf("the kind changes:\n%s", storage)
	}
	if !strings.Contains(storage, "'../../libs/util/main.bicep'") {
		t.Fatalf("a module it used beside it is reached through libs/ now:\n%s", storage)
	}
	if !strings.Contains(got["apps/infra/main.bicep"], "'../storage/main.bicep'") {
		t.Fatalf("dependents follow:\n%s", got["apps/infra/main.bicep"])
	}
	if _, touched := got["libs/util/main.bicep"]; touched {
		t.Fatal("a project that does not use storage is not edited")
	}
}

func TestAPackageIsRefused(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{Project: projects()[0], To: protocol.ProjectPackage, NewRoot: "packages/storage", Projects: projects()})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Refused, "registry") || len(r.Files) != 0 {
		t.Fatalf("refused, naming why, and nothing is edited: %+v", r)
	}
}

func TestMovingToADifferentDepthRebasesItsOwnModules(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{Project: projects()[0], To: protocol.ProjectLib, NewRoot: "libs/group/storage", Projects: projects()})
	if err != nil {
		t.Fatal(err)
	}
	if got := edited(r)["libs/storage/main.bicep"]; !strings.Contains(got, "'../../util/main.bicep'") {
		t.Fatalf("one level deeper:\n%s", got)
	}
}

func TestLibToAppAndBackLeavesTheFileByteForByte(t *testing.T) {
	dir := writeFiles(t, map[string]string{"libs/storage/main.bicep": storageBicep})
	project := protocol.Project{Name: "storage", Root: "libs/storage", Language: "bicep"}
	for _, to := range []protocol.ProjectKind{protocol.ProjectApp, protocol.ProjectLib} {
		r, err := Change(dir, protocol.ChangeKindParams{Project: project, To: to, NewRoot: "libs/storage"})
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range r.Files {
			if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(e.Path)), []byte(e.Content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "libs/storage/main.bicep")); string(got) != storageBicep {
		t.Fatalf("round trip changed the file:\n%s", got)
	}
}

func TestAppToLibKeepsItsParametersAndSaysSo(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{Project: projects()[2], To: protocol.ProjectLib, NewRoot: "libs/infra", Projects: projects()})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "main.bicepparam") {
		t.Fatalf("the parameters file is named: %v", r.Warnings)
	}
}
