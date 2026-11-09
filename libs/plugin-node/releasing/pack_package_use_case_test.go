package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTarballName(t *testing.T) {
	for name, want := range map[string]string{"@acme/sdk": "acme-sdk-1.2.0.tgz", "plain": "plain-1.2.0.tgz"} {
		if got := TarballName(name, "1.2.0"); got != want {
			t.Errorf("%s: got %s, want %s", name, got, want)
		}
	}
}

func TestPackWritesTheTarballAndPublishesNothing(t *testing.T) {
	root := t.TempDir()
	calls := fakeNpm(t, "")
	npm.run = func(_ string, argv []string) error {
		*calls = append(*calls, strings.Join(argv, " "))
		dir := filepath.Join(root, "dist", "release", "sdk", "packages")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "acme-sdk-1.1.0.tgz"), []byte("tgz"), 0o644)
	}
	params := protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk", Exports: []string{"@acme/sdk"}}, Version: "1.1.0", PackDir: "dist/release/sdk/packages"}
	res, err := Publish(root, params)
	if err != nil {
		t.Fatal(err)
	}
	if res.Published || len(res.Assets) != 1 || res.Assets[0] != "dist/release/sdk/packages/acme-sdk-1.1.0.tgz" {
		t.Fatalf("result %+v", res)
	}
	for _, c := range *calls {
		if !strings.HasPrefix(c, "npm pack ") {
			t.Fatalf("pack mode runs only npm pack, ran %q", c)
		}
	}
}

func TestPackDryRunNamesTheFileAndRunsNothing(t *testing.T) {
	calls := fakeNpm(t, "")
	res, err := Publish(t.TempDir(), protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk", Exports: []string{"@acme/sdk"}}, Version: "1.1.0", DryRun: true, PackDir: "out"})
	if err != nil || res.Skipped != "dry run" || len(res.Assets) != 1 || len(*calls) != 0 {
		t.Fatalf("res %+v err %v calls %v", res, err, *calls)
	}
}

func TestPackFailsWhenNpmWritesNothing(t *testing.T) {
	fakeNpm(t, "")
	_, err := Publish(t.TempDir(), protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk", Exports: []string{"@acme/sdk"}}, Version: "1.1.0", PackDir: "out"})
	if err == nil {
		t.Fatal("a missing tarball must fail the release")
	}
}
