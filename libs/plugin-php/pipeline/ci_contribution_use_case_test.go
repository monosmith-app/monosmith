package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCI(t *testing.T) {
	ws := t.TempDir()
	res := CI(ws, protocol.CIParams{Projects: []protocol.Project{{Kind: protocol.ProjectPackage}}})
	if res.Toolchains[0].Version != PHPVersion || res.Setup[0].Command[1] != "install" || len(res.Credentials) != 0 {
		t.Fatalf("%+v", res)
	}
	_ = os.WriteFile(filepath.Join(ws, ".php-version"), []byte("8.3\n"), 0o644)
	if v := Version(ws); v != "8.3" {
		t.Fatalf(".php-version pins %s", v)
	}
}

func TestDoctor(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(_ string, argv []string) (string, error) {
		if argv[0] == "composer" {
			return "", errors.New("missing")
		}
		return "", nil
	}
	res := Doctor(t.TempDir(), protocol.DoctorParams{Projects: []protocol.Project{{}}})
	if len(res.Findings) != 2 || !res.Findings[0].OK || res.Findings[1].OK {
		t.Fatalf("%+v", res)
	}
}

func TestCIDeclaresTheComposerDownloadStore(t *testing.T) {
	got := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{{Name: "lib", Root: "packages/lib"}}}).Stores
	want := []protocol.PackageStore{{Name: "composer", Path: "~/.cache/composer", Env: "COMPOSER_CACHE_DIR", Files: []string{"**/composer.lock"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%+v", got)
	}
}
