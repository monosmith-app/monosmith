package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var (
	core = protocol.Project{Kind: protocol.ProjectPackage, Tasks: map[string]protocol.Task{"lint": {Command: []string{"dart", "analyze"}}}}
	web  = protocol.Project{Kind: protocol.ProjectApp, Tasks: map[string]protocol.Task{"lint": {Command: []string{"flutter", "analyze"}}}}
)

func TestCI(t *testing.T) {
	ws := t.TempDir()
	res := CI(ws, protocol.CIParams{Projects: []protocol.Project{core}})
	if res.Toolchains[0].Name != "dart" || res.Setup[0].Command[0] != "dart" || res.Credentials[0].Env != "PUB_TOKEN" || res.Credentials[0].Required {
		t.Fatalf("%+v", res)
	}
	_ = os.WriteFile(filepath.Join(ws, ".flutter-version"), []byte("3.44.0\n"), 0o644)
	res = CI(ws, protocol.CIParams{Projects: []protocol.Project{core, web}})
	if tc := res.Toolchains[0]; tc.Name != "flutter" || tc.Version != "3.44.0" || res.Setup[0].Command[0] != "flutter" {
		t.Fatalf("%+v", res)
	}
}

func TestDoctor(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(_ string, argv []string) (string, error) {
		if argv[0] == "flutter" {
			return "", errors.New("missing")
		}
		return "", nil
	}
	ws := t.TempDir()
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{core, web}})
	if len(res.Findings) != 3 || !res.Findings[0].OK || res.Findings[1].OK || !res.Findings[2].OK {
		t.Fatalf("%+v", res)
	}
	_ = os.Mkdir(filepath.Join(ws, "flutter"), 0o755)
	if res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{core}}); res.Findings[1].OK {
		t.Fatalf("an SDK inside the workspace fails: %+v", res)
	}
}

func TestCIDeclaresThePubDownloadStore(t *testing.T) {
	got := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{{Name: "app", Root: "apps/app"}}}).Stores
	want := []protocol.PackageStore{{Name: "pub", Path: "~/.pub-cache", Env: "PUB_CACHE", Files: []string{"**/pubspec.lock"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%+v", got)
	}
}
