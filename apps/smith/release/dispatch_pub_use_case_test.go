package release

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func pubApp(prefix, pattern string) *projectgraph.Project {
	return &projectgraph.Project{
		Project: protocol.Project{Name: "smith", Root: "apps/smith", Kind: protocol.ProjectApp, TagPrefix: prefix},
		Release: &workspace.ProjectRelease{Pub: "monosmith", PubTag: pattern},
	}
}

func TestPubTagSpellsTheReleaseTagWithoutTheAt(t *testing.T) {
	cases := []struct {
		p    *projectgraph.Project
		want string
	}{
		{pubApp("", ""), "smith-v1.4.0"},
		{pubApp("cli@", ""), "cli-v1.4.0"},
		{pubApp("", "v{version}"), "v1.4.0"},
		{pubApp("", "{name}/{version}"), "smith/1.4.0"},
		// A prefix that already ends in -v is the pub tag's form: the release tag is the pub tag.
		{pubApp("cli-v", ""), "cli-v1.4.0"},
		{pubApp("v", ""), "v-v1.4.0"},
	}
	for _, c := range cases {
		if got := PubTag(c.p, "1.4.0"); got != c.want {
			t.Errorf("PubTag(%s, %q) = %s, want %s", c.p.TagPrefix, c.p.Release.PubTag, got, c.want)
		}
	}
}

// released is a workspace whose smith@1.4.0 is tagged and pushed to a bare
// origin, as the release leaves it before dispatchPub.
func released(t *testing.T) *session.Session {
	t.Helper()
	g := repo(t)
	commit(t, g, "apps/smith/main.go", "feat: smith")
	origin := gitclient.New(t.TempDir())
	if _, err := origin.Run("init", "--bare"); err != nil {
		t.Skip("git unavailable")
	}
	if _, err := g.Run("remote", "add", "origin", origin.Dir); err != nil {
		t.Fatal(err)
	}
	if err := g.CreateTag("smith@1.4.0", "smith 1.4.0"); err != nil {
		t.Fatal(err)
	}
	commit(t, g, "README.md", "docs: later")
	cfg := workspace.Config{}
	cfg.Release.Remote, cfg.Release.GitHubReleases = "origin", true
	return &session.Session{Root: g.Dir, Config: cfg, Git: g}
}

func TestDispatchPubTagsTheReleaseCommitAndStartsTheWorkflow(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "acme/acme")
	t.Setenv("TF_BUILD", "")
	t.Setenv("GITLAB_CI", "")
	s := released(t)
	keep, keepHas := dispatcher, pubHas
	defer func() { dispatcher, pubHas = keep, keepHas }()
	var ref string
	var inputs map[string]string
	dispatcher = func(_ githubrelease.Repository, r string, in map[string]string) error {
		ref, inputs = r, in
		return nil
	}
	pubHas = func(string, string) (bool, error) { return false, nil }

	item := Item{Project: "smith", Next: "1.4.0", Tag: "smith@1.4.0"}
	res := ReportResult{Pushed: true}
	if err := dispatchPub(s, pubApp("", ""), item, &res, func(string, ...any) {}); err != nil {
		t.Fatal(err)
	}
	if got, want := s.Git.TagCommit("smith-v1.4.0"), s.Git.TagCommit("smith@1.4.0"); got == "" || got != want {
		t.Errorf("smith-v1.4.0 is on %q, the release on %q", got, want)
	}
	if out, _ := s.Git.Run("ls-remote", "--tags", "origin", "smith-v1.4.0"); out == "" {
		t.Error("the pub tag was not pushed")
	}
	if ref != "refs/tags/smith-v1.4.0" || inputs["project"] != "smith" || inputs["version"] != "1.4.0" {
		t.Errorf("dispatched %s with %v", ref, inputs)
	}
	if res.PubDispatch != "monosmith 1.4.0 from smith-v1.4.0" {
		t.Errorf("report = %q", res.PubDispatch)
	}

	// A re-run keeps the tag and dispatches again while pub.dev lacks it.
	ref = ""
	if err := dispatchPub(s, pubApp("", ""), item, &res, func(string, ...any) {}); err != nil || ref == "" {
		t.Errorf("re-run: dispatched %q, %v", ref, err)
	}
}

func TestDispatchPubSkipsWhatPubDevHasAndOtherCIs(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "acme/acme")
	t.Setenv("GITLAB_CI", "")
	s := released(t)
	keep, keepHas := dispatcher, pubHas
	defer func() { dispatcher, pubHas = keep, keepHas }()
	dispatched := false
	dispatcher = func(githubrelease.Repository, string, map[string]string) error { dispatched = true; return nil }
	pubHas = func(string, string) (bool, error) { return true, nil }
	item := Item{Project: "smith", Next: "1.4.0", Tag: "smith@1.4.0"}

	t.Setenv("TF_BUILD", "")
	res := ReportResult{Pushed: true}
	if err := dispatchPub(s, pubApp("", ""), item, &res, func(string, ...any) {}); err != nil {
		t.Fatal(err)
	}
	if dispatched || len(res.ExistingPackages) != 1 || res.ExistingPackages[0] != "monosmith@1.4.0" {
		t.Errorf("pub.dev has it: dispatched %v, existing %v", dispatched, res.ExistingPackages)
	}
	if s.Git.TagCommit("smith-v1.4.0") != "" {
		t.Error("tagged a version pub.dev has")
	}

	pubHas = func(string, string) (bool, error) { return false, nil }
	t.Setenv("TF_BUILD", "True")
	if err := dispatchPub(s, pubApp("", ""), item, &ReportResult{Pushed: true}, func(string, ...any) {}); err != nil || dispatched {
		t.Errorf("Azure Pipelines: dispatched %v, %v", dispatched, err)
	}
	t.Setenv("TF_BUILD", "")
	if err := dispatchPub(s, pubApp("", ""), item, &ReportResult{}, func(string, ...any) {}); err != nil || dispatched {
		t.Errorf("tag not pushed: dispatched %v, %v", dispatched, err)
	}
}

func TestPubPackageNamesWhatPublishesToPubDev(t *testing.T) {
	pkg := func(publishTo string, exports ...string) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: "core", Kind: protocol.ProjectPackage, PublishTo: publishTo, Exports: exports}}
	}
	cases := []struct {
		name string
		p    *projectgraph.Project
		want string
	}{
		{"an app's configured package", pubApp("", ""), "monosmith"},
		{"a dart package that publishes to pub.dev", pkg("pub.dev", "acme_core"), "acme_core"},
		{"a package with a private repository", pkg("https://pub.example", "acme_core"), ""},
		{"a package that never publishes", pkg("none", "acme_core"), ""},
		{"a package that says nothing", pkg("", "acme_core"), ""},
		{"a package without an export", pkg("pub.dev"), ""},
	}
	for _, c := range cases {
		if got := PubPackage(c.p); got != c.want {
			t.Errorf("%s: PubPackage = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestDispatchPubStartsThePubWorkflowForAPackage(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "acme/acme")
	t.Setenv("TF_BUILD", "")
	t.Setenv("GITLAB_CI", "")
	s := released(t)
	keep, keepHas := dispatcher, pubHas
	defer func() { dispatcher, pubHas = keep, keepHas }()
	var ref string
	dispatcher = func(_ githubrelease.Repository, r string, _ map[string]string) error {
		ref = r
		return nil
	}
	pubHas = func(string, string) (bool, error) { return false, nil }

	core := &projectgraph.Project{Project: protocol.Project{Name: "smith", Kind: protocol.ProjectPackage, PublishTo: "pub.dev", Exports: []string{"acme_core"}}}
	res := ReportResult{Pushed: true}
	item := Item{Project: "smith", Next: "1.4.0", Tag: "smith@1.4.0"}
	if err := dispatchPub(s, core, item, &res, func(string, ...any) {}); err != nil {
		t.Fatal(err)
	}
	if ref != "refs/tags/smith-v1.4.0" || !strings.Contains(res.PubDispatch, "acme_core 1.4.0") {
		t.Errorf("ref %q, dispatch %q", ref, res.PubDispatch)
	}
}

// A pub tag left on an older commit (a changed tag_prefix, a hand-made tag)
// would start the pub run from that commit, with its old workflow and smith.
func TestDispatchPubStopsWhenThePubTagIsOnAnotherCommit(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "acme/acme")
	t.Setenv("TF_BUILD", "")
	t.Setenv("GITLAB_CI", "")
	s := released(t)
	keep, keepHas := dispatcher, pubHas
	defer func() { dispatcher, pubHas = keep, keepHas }()
	dispatched := false
	dispatcher = func(githubrelease.Repository, string, map[string]string) error { dispatched = true; return nil }
	pubHas = func(string, string) (bool, error) { return false, nil }
	if err := s.Git.CreateTag("smith-v1.4.0", "old"); err != nil {
		t.Fatal(err)
	}
	// Move the release tag's commit away from the pub tag's.
	if _, err := s.Git.Run("tag", "-d", "smith@1.4.0"); err != nil {
		t.Fatal(err)
	}
	commit(t, s.Git, "other.txt", "fix: later")
	if err := s.Git.CreateTag("smith@1.4.0", "smith 1.4.0"); err != nil {
		t.Fatal(err)
	}
	item := Item{Project: "smith", Next: "1.4.0", Tag: "smith@1.4.0"}
	err := dispatchPub(s, pubApp("", ""), item, &ReportResult{Pushed: true}, func(string, ...any) {})
	if err == nil || !strings.Contains(err.Error(), "smith-v1.4.0 already exists on") || dispatched {
		t.Errorf("err %v, dispatched %v", err, dispatched)
	}
}

// With a tag prefix like acme-v the release tag is already the pub tag: no
// second tag is made, and the workflow starts on the release tag.
func TestDispatchPubUsesTheReleaseTagWhenItIsAlreadyThePubTag(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "acme/acme")
	t.Setenv("TF_BUILD", "")
	t.Setenv("GITLAB_CI", "")
	s := released(t)
	keep, keepHas := dispatcher, pubHas
	defer func() { dispatcher, pubHas = keep, keepHas }()
	var ref string
	dispatcher = func(_ githubrelease.Repository, r string, _ map[string]string) error { ref = r; return nil }
	pubHas = func(string, string) (bool, error) { return false, nil }
	if err := s.Git.CreateTag("smith-v1.4.0", "smith 1.4.0"); err != nil {
		t.Fatal(err)
	}
	item := Item{Project: "smith", Next: "1.4.0", Tag: "smith-v1.4.0"}
	if err := dispatchPub(s, pubApp("smith-v", ""), item, &ReportResult{Pushed: true}, func(string, ...any) {}); err != nil {
		t.Fatal(err)
	}
	if ref != "refs/tags/smith-v1.4.0" {
		t.Errorf("dispatched %q", ref)
	}
	if s.Git.TagCommit("smith-v-v1.4.0") != "" {
		t.Error("a second, doubled pub tag was made")
	}
}
