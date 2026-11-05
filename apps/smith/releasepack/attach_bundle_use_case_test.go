package releasepack

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTheBundleGoesOnTheReleaseOfEveryPackageAndAReRunKeepsTheOthers(t *testing.T) {
	fake := newFake("core@0.5.2", "cli@0.2.4")
	srv := httptest.NewServer(fake)
	defer srv.Close()
	c := client(srv)
	root := t.TempDir()

	// A first run that finished only core.
	if _, err := AttachBundle(c, repo, root, NewBundle("o/r", []Package{npmPackage("core", "0.5.2", "core-0.5.2.tgz")})); err != nil {
		t.Fatal(err)
	}
	// The re-run finishes cli and does not know core: the bundle keeps core.
	tags, err := AttachBundle(c, repo, root, NewBundle("o/r", []Package{npmPackage("cli", "0.2.4", "cli-0.2.4.tgz"), npmPackage("core", "0.5.2", "core-0.5.2.tgz")}))
	if err != nil || len(tags) != 2 {
		t.Fatalf("%v %v", tags, err)
	}
	for _, tag := range []string{"core@0.5.2", "cli@0.2.4"} {
		names := fake.assets(tag)
		if len(names) != 2 || !strings.Contains(strings.Join(names, ","), BundleFile) || !strings.Contains(strings.Join(names, ","), BundleMarkdown) {
			t.Fatalf("%s holds %v", tag, names)
		}
		var b Bundle
		if err := json.Unmarshal([]byte(fake.body(tag, BundleFile)), &b); err != nil || len(b.Packages) != 2 {
			t.Fatalf("%s: both packages are in the bundle: %v %+v", tag, err, b)
		}
	}
	// A run that adds nothing changes nothing.
	if _, err := AttachBundle(c, repo, root, NewBundle("o/r", []Package{npmPackage("cli", "0.2.4", "cli-0.2.4.tgz"), npmPackage("core", "0.5.2", "core-0.5.2.tgz")})); err != nil {
		t.Fatal(err)
	}
	if len(fake.assets("core@0.5.2")) != 2 {
		t.Fatalf("the same bundle is not attached twice: %v", fake.assets("core@0.5.2"))
	}
}

func TestABundleNeedsTheReleaseToExist(t *testing.T) {
	srv := httptest.NewServer(newFake())
	defer srv.Close()
	if _, err := AttachBundle(client(srv), repo, t.TempDir(), NewBundle("o/r", []Package{npmPackage("core", "0.5.2", "core-0.5.2.tgz")})); err == nil {
		t.Fatal("a missing release is an error")
	}
}
