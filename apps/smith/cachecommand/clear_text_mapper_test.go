package cachecommand

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/remotesync"
)

func TestTheRemoteClearQuestionNamesTheStoreTheWorkspaceAndTheCount(t *testing.T) {
	loc := location{Dir: "/w/.smith/cache", Workspace: "acme"}
	listed := remotesync.RemoteObjects{Names: []string{"ac/1", "cas/2", "cas/3"}, Listed: true}
	got := clearQuestion(loc, "s3://bucket/team/cache?x=1", listed, false, cachemaintenance.Status{})
	want := "Delete 3 objects of the remote cache s3://bucket/team/cache for the workspace acme?\nThe team and CI share the remote cache: their next runs build again what it held."
	if got != want {
		t.Errorf("remote:\n%s\nwant:\n%s", got, want)
	}
	got = clearQuestion(loc, "https://cache.example.com/team", remotesync.RemoteObjects{Names: []string{"ac/1"}}, true, cachemaintenance.Status{Exists: true, Entries: 4, Bytes: 2048})
	for _, w := range []string{"Delete the local cache at /w/.smith/cache (4 entries, 2.0 KB) and up to 1 object of the remote cache https://cache.example.com/team for the workspace acme?", "This store cannot list its objects: smith deletes only the entries this machine's cache holds"} {
		if !strings.Contains(got, w) {
			t.Errorf("all, plain store: missing %q in:\n%s", w, got)
		}
	}
}

func TestClearRemoteAsksDeletesAndLeavesTheLocalCache(t *testing.T) {
	_, store := workspaceWithCache(t, "", 2, 100)
	dir := filepath.Dir(store.Root())
	srv, _ := objectServer(t)
	withRemote(t, srv, "team-token")
	if _, err := run(t, "push"); err != nil {
		t.Fatal(err)
	}

	out, err := run(t, "clear", "--remote", "--dry-run")
	if err != nil || !strings.Contains(out, "Delete up to 6 objects of the remote cache") || !strings.Contains(out, "for the workspace acme?") || !strings.Contains(out, "Dry run: nothing was deleted.") {
		t.Fatalf("dry run: %v\n%s", err, out)
	}
	_, err = run(t, "clear", "--remote")
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Problem, "without a terminal") || !strings.Contains(strings.Join(p.Remedy, " "), "smith cache clear --remote --yes") {
		t.Fatalf("without a terminal: %v", err)
	}
	out, err = run(t, "clear", "--remote", "--yes")
	if err != nil || !strings.Contains(out, "✓ Deleted 6 of 6 objects of the remote cache") || !strings.Contains(out, "This store cannot list") {
		t.Fatalf("clear --remote: %v\n%s", err, out)
	}
	if status, _ := cachemaintenance.Inspect(dir, 0); status.Entries != 2 {
		t.Fatal("--remote must leave the local cache")
	}
	out, err = run(t, "push", "--dry-run")
	if err != nil || !strings.Contains(out, "2 of 2 entries would be uploaded") {
		t.Fatalf("the remote must be empty again: %v\n%s", err, out)
	}

	out, err = run(t, "clear", "--all", "--yes")
	if err != nil || !strings.Contains(out, "Deleted the local cache at "+dir) || !strings.Contains(out, "of the remote cache") {
		t.Fatalf("clear --all: %v\n%s", err, out)
	}
	if _, err := os.Stat(store.Root()); err == nil {
		t.Fatal("--all must delete the local cache")
	}
}

func TestClearRemoteRefusedSaysTheRightsItNeeds(t *testing.T) {
	workspaceWithCache(t, "", 1, 10)
	srv, _ := objectServer(t)
	withRemote(t, srv, "old-token")
	_, err := run(t, "clear", "--remote", "--yes")
	if err == nil || !strings.Contains(err.Error(), "refused the credentials") || !strings.Contains(err.Error(), "list and delete rights") {
		t.Fatalf("err = %v", err)
	}
}

func TestClearRemoteAndAllTogetherIsRefused(t *testing.T) {
	workspaceWithCache(t, "", 1, 10)
	if _, err := run(t, "clear", "--remote", "--all"); err == nil {
		t.Fatal("--remote --all must be refused")
	}
}
