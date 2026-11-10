package releasing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPackCopiesTheWheelAndTheSdistWithoutTwine(t *testing.T) {
	root := t.TempDir()
	built := filepath.Join(root, "dist", "packages", "py")
	if err := os.MkdirAll(built, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"acme-1.0.0-py3-none-any.whl", "acme-1.0.0.tar.gz", "acme-0.9.0.tar.gz"} {
		if err := os.WriteFile(filepath.Join(built, n), []byte(n), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	Twine = func(string, []string, []string) (string, error) {
		t.Fatal("pack mode never runs twine")
		return "", nil
	}
	res, err := Publish(root, protocol.PublishParams{Project: protocol.Project{Root: "packages/py", Exports: []string{"acme"}}, Version: "1.0.0", PackDir: "dist/release/py/packages"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Published || len(res.Assets) != 2 || res.Assets[0] != "dist/release/py/packages/acme-1.0.0-py3-none-any.whl" || res.Assets[1] != "dist/release/py/packages/acme-1.0.0.tar.gz" {
		t.Fatalf("result %+v", res)
	}
	if _, err := os.Stat(filepath.Join(root, "dist", "release", "py", "packages", "acme-1.0.0.tar.gz")); err != nil {
		t.Fatal(err)
	}
}

func TestPackFailsWithoutABuild(t *testing.T) {
	_, err := Publish(t.TempDir(), protocol.PublishParams{Project: protocol.Project{Root: "packages/py"}, Version: "1.0.0", PackDir: "out"})
	if err == nil {
		t.Fatal("a release with no distribution must fail")
	}
}

func TestPackDryRunWritesNothing(t *testing.T) {
	root := t.TempDir()
	res, err := Publish(root, protocol.PublishParams{Project: protocol.Project{Root: "packages/py"}, Version: "1.0.0", DryRun: true, PackDir: "out"})
	if err != nil || res.Skipped != "dry run" {
		t.Fatalf("res %+v err %v", res, err)
	}
	if _, err := os.Stat(filepath.Join(root, "out")); err == nil {
		t.Fatal("a dry run writes nothing")
	}
}
