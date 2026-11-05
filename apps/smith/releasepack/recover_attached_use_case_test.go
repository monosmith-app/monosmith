package releasepack

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
)

func TestACompleteReleaseIsRecoveredFromItsChecksumFiles(t *testing.T) {
	fake := newFake("core@0.5.2")
	srv := httptest.NewServer(fake)
	defer srv.Close()
	c := client(srv)
	dir := t.TempDir()
	pkg := filepath.Join(dir, "core-0.5.2.tgz")
	_ = os.WriteFile(pkg, []byte("abc"), 0o644)
	files, sidecars, _ := Checksum(dir, []string{"core-0.5.2.tgz"})
	eco, _ := EcosystemOf("node")

	if _, err := githubrelease.Ensure(c, repo, githubrelease.Request{Tag: "core@0.5.2", Assets: []string{pkg}}); err != nil {
		t.Fatal(err)
	}
	rel, _, _ := c.ReleaseByTag(repo, "core@0.5.2")
	if Complete(rel, eco) {
		t.Fatal("a package file without its checksum file is an attach that stopped")
	}
	if _, err := githubrelease.Ensure(c, repo, githubrelease.Request{Tag: "core@0.5.2", Assets: []string{filepath.Join(dir, sidecars[0])}}); err != nil {
		t.Fatal(err)
	}
	rel, _, _ = c.ReleaseByTag(repo, "core@0.5.2")
	if !Complete(rel, eco) {
		t.Fatal("the package file and its checksum file are there")
	}
	got, err := Recover(c, repo, rel, eco)
	if err != nil || len(got) != 1 || got[0].SHA256 != files[0].SHA256 || got[0].Size != 3 {
		t.Fatalf("%+v %v", got, err)
	}
}
