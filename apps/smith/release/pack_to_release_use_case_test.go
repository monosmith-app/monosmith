package release

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/releasepack"
)

func tgz(project, version, salt string) map[string]string {
	return map[string]string{project + "-" + version + ".tgz": "tarball of " + project + " " + version + salt}
}

func wheels(project, version, salt string) map[string]string {
	return map[string]string{
		project + "-" + version + "-py3-none-any.whl": "wheel of " + project + salt,
		project + "-" + version + ".tar.gz":           "sdist of " + project + salt,
	}
}

var npmPlan = Plan{Items: []Item{
	{Project: "core", Plugin: "node", Current: "0.5.1", Next: "0.5.2", Tag: "core@0.5.2", Level: "patch", Reason: "fix: x"},
	{Project: "cli", Plugin: "node", Current: "0.2.3", Next: "0.2.4", Tag: "cli@0.2.4", Level: "patch", Reason: "dependency core 0.5.2"},
}}

func TestARealReleasePacksAttachesChecksumsAfterTheFilesAndPublishesToNoRegistry(t *testing.T) {
	gh := newFakeGitHub(t)
	s, plugin := packWorkspace(t, "python", wheels, "core")
	plan := Plan{Items: []Item{{Project: "core", Plugin: "python", Current: "0.5.1", Next: "0.5.2", Tag: "core@0.5.2", Level: "patch", Reason: "fix: x"}}}
	var out bytes.Buffer
	report, err := Execute(context.Background(), s, plan, Options{Push: true, Out: &out})
	if err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	want := []string{"core@0.5.2/core-0.5.2-py3-none-any.whl", "core@0.5.2/core-0.5.2.tar.gz", "core@0.5.2/core-0.5.2-py3-none-any.whl.sha256", "core@0.5.2/core-0.5.2.tar.gz.sha256", "core@0.5.2/release-bundle.json", "core@0.5.2/INSTALL.md"}
	if strings.Join(gh.uploads, "\n") != strings.Join(want, "\n") {
		t.Fatalf("every package file, then every checksum file, then the bundle:\n%v", gh.uploads)
	}
	if packs, publishes := plugin.counts(); packs != 1 || publishes != 0 {
		t.Fatalf("packed once and never published to a registry: %d %d", packs, publishes)
	}
	if plugin.packDirs[0] != "dist/release/core/packages" {
		t.Fatalf("pack folder %v", plugin.packDirs)
	}
	res := report.Results[0]
	if res.Status != "released" || res.Publish != attachedToRelease || !res.Pushed || len(res.Packed) != 2 || res.Published {
		t.Fatalf("%+v", res)
	}
	if got := gh.body("core@0.5.2", "core-0.5.2.tar.gz.sha256"); got != sha("sdist of core")+"  core-0.5.2.tar.gz\n" || res.Packed[1].SHA256 != sha("sdist of core") {
		t.Fatalf("the checksum file states the digest of what was attached: %q", got)
	}
	if s.Git.TagCommit("core@0.5.2") == "" {
		t.Fatal("the tag exists")
	}
	if journal, _ := ReadJournal(s.Root, "core"); !strings.Contains(string(mustJSON(t, journal)), "core-0.5.2.tar.gz sha256:"+sha("sdist of core")) {
		t.Fatalf("the journal records the asset names: %+v", journal)
	}
	for _, line := range gh.log {
		if !strings.Contains(line, "/repos/o/r/") && !strings.Contains(line, "/upload/") {
			t.Fatalf("only GitHub is asked: %s", line)
		}
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestTheBundleIsOnTheReleaseOfEveryPackageOfTheRun(t *testing.T) {
	gh := newFakeGitHub(t)
	s, _ := packWorkspace(t, "node", tgz, "cli", "core")
	var out bytes.Buffer
	if _, err := Execute(context.Background(), s, npmPlan, Options{Push: true, Out: &out}); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	for _, tag := range []string{"core@0.5.2", "cli@0.2.4"} {
		var b releasepack.Bundle
		if err := json.Unmarshal([]byte(gh.body(tag, releasepack.BundleFile)), &b); err != nil || len(b.Packages) != 2 || b.Repository != "o/r" {
			t.Fatalf("%s: both packages are in the bundle: %v %+v", tag, err, b)
		}
		install := gh.body(tag, releasepack.BundleMarkdown)
		for _, want := range []string{"npm i ./cli-0.2.4.tgz ./core-0.5.2.tgz", `gh release download "core@0.5.2" --repo o/r`, `gh release download "cli@0.2.4" --repo o/r`, "gh auth login", sha("tarball of core 0.5.2")} {
			if !strings.Contains(install, want) {
				t.Errorf("%s INSTALL.md lacks %q:\n%s", tag, want, install)
			}
		}
	}
	if !strings.Contains(out.String(), "attached release-bundle.json and INSTALL.md to 2 GitHub Releases") {
		t.Fatalf("the run says so:\n%s", out.String())
	}
}

func TestAFailedAttachResumesAndAttachesOnlyWhatIsMissing(t *testing.T) {
	gh := newFakeGitHub(t)
	s, plugin := packWorkspace(t, "node", tgz, "core")
	plan := Plan{Items: npmPlan.Items[:1]}
	gh.failUpload = func(name string) bool { return strings.HasSuffix(name, ".sha256") }
	var out bytes.Buffer
	if _, err := Execute(context.Background(), s, plan, Options{Push: true, Out: &out}); err == nil {
		t.Fatal("the failed upload fails the release")
	}
	if names := gh.assetNames("core@0.5.2"); len(names) != 1 || names[0] != "core-0.5.2.tgz" {
		t.Fatalf("the file is attached, its checksum is not: %v", names)
	}
	if s.Git.TagCommit("core@0.5.2") == "" {
		t.Fatal("the tag stays, as for any release that stopped after tagging")
	}
	items, err := UnfinishedPacks(s, func(string, releasepack.Ecosystem) (bool, error) { return false, nil }, &out)
	if err != nil || len(items) != 1 || items[0].Project != "core" || items[0].Reason != PackResumeReason {
		t.Fatalf("a CI re-run on a fresh checkout plans the unfinished package: %+v %v", items, err)
	}
	gh.uploads = nil
	report, err := Execute(context.Background(), s, plan, Options{Push: true, Out: &out})
	if err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if want := []string{"core@0.5.2/core-0.5.2.tgz.sha256", "core@0.5.2/release-bundle.json", "core@0.5.2/INSTALL.md"}; strings.Join(gh.uploads, ",") != strings.Join(want, ",") || gh.deleted != 0 {
		t.Fatalf("the identical file is kept, only the missing assets go up: %v (deleted %d)", gh.uploads, gh.deleted)
	}
	if res := report.Results[0]; res.Status != "already-released" || len(res.Packed) != 1 || res.Publish != attachedToRelease {
		t.Fatalf("%+v", res)
	}
	if _, publishes := plugin.counts(); publishes != 0 {
		t.Fatal("the resume never publishes to a registry")
	}
}

func TestAResumedFileThatDiffersIsReplacedAndItsChecksumDescribesTheNewOne(t *testing.T) {
	gh := newFakeGitHub(t)
	s, plugin := packWorkspace(t, "node", tgz, "core")
	plan := Plan{Items: npmPlan.Items[:1]}
	gh.failUpload = func(name string) bool { return strings.HasSuffix(name, ".sha256") }
	var out bytes.Buffer
	if _, err := Execute(context.Background(), s, plan, Options{Push: true, Out: &out}); err == nil {
		t.Fatal("the failed upload fails the release")
	}
	// Packed again, the file has the same size and other bytes, as a wheel does.
	plugin.salt = "!"
	if _, err := Execute(context.Background(), s, plan, Options{Push: true, Out: &out}); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if gh.deleted != 1 {
		t.Fatalf("the differing file is replaced: deleted %d", gh.deleted)
	}
	want := "tarball of core 0.5.2!"
	if gh.body("core@0.5.2", "core-0.5.2.tgz") != want || gh.body("core@0.5.2", "core-0.5.2.tgz.sha256") != sha(want)+"  core-0.5.2.tgz\n" {
		t.Fatal("the attached file and its checksum file agree")
	}
}

func TestACompleteReleaseIsReadBackNotPackedAgain(t *testing.T) {
	gh := newFakeGitHub(t)
	s, plugin := packWorkspace(t, "node", tgz, "core")
	plan := Plan{Items: npmPlan.Items[:1]}
	var out bytes.Buffer
	if _, err := Execute(context.Background(), s, plan, Options{Push: true, Out: &out}); err != nil {
		t.Fatal(err)
	}
	gh.uploads = nil
	report, err := Execute(context.Background(), s, plan, Options{Push: true, Out: &out})
	if err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if packs, _ := plugin.counts(); packs != 1 {
		t.Fatalf("not packed again: %d", packs)
	}
	if len(gh.uploads) != 0 || gh.deleted != 0 {
		t.Fatalf("nothing uploaded: %v", gh.uploads)
	}
	if res := report.Results[0]; len(res.Packed) != 1 || res.Packed[0].SHA256 != sha("tarball of core 0.5.2") || res.Packed[0].Size == 0 {
		t.Fatalf("the digest is read back from the checksum file: %+v", res.Packed)
	}
}

func TestAReleaseThatLacksFilesAndIsBehindHeadIsNotPackedFromNewerCode(t *testing.T) {
	gh := newFakeGitHub(t)
	s, _ := packWorkspace(t, "node", tgz, "core")
	plan := Plan{Items: npmPlan.Items[:1]}
	gh.failUpload = func(name string) bool { return strings.HasSuffix(name, ".sha256") }
	var out bytes.Buffer
	_, _ = Execute(context.Background(), s, plan, Options{Push: true, Out: &out})
	commit(t, s.Git, "packages/core/index.js", "feat: later")
	_, err := Execute(context.Background(), s, plan, Options{Push: true, Out: &out})
	if err == nil || !strings.Contains(err.Error(), "HEAD is newer than its tag") {
		t.Fatalf("%v", err)
	}
}

func TestADryRunListsTheFilesAndTheBundleAndAsksGitHubNothing(t *testing.T) {
	gh := newFakeGitHub(t)
	s, plugin := packWorkspace(t, "node", tgz, "cli", "core")
	var out bytes.Buffer
	if _, err := Execute(context.Background(), s, npmPlan, Options{DryRun: true, Out: &out}); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	for _, want := range []string{"would attach dist/release/core/packages/core-0.5.2.tgz", "would attach dist/release/cli/packages/cli-0.2.4.tgz", "would attach release-bundle.json and INSTALL.md to the GitHub Release of each package"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the dry run lacks %q:\n%s", want, out.String())
		}
	}
	if gh.requests() != 0 {
		t.Fatalf("a dry run asks GitHub nothing: %v", gh.log)
	}
	if packs, publishes := plugin.counts(); packs != 0 || publishes != 0 {
		t.Fatal("a dry run writes and publishes nothing")
	}
	if s.Git.TagCommit("core@0.5.2") != "" {
		t.Fatal("a dry run tags nothing")
	}
}

// writeTemp writes a file in a new folder and returns its path.
func writeTemp(t *testing.T, name string, content []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, content, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestAnExistingBundleIsMergedNotReplacedByASmallerOne(t *testing.T) {
	gh := newFakeGitHub(t)
	s, _ := packWorkspace(t, "node", tgz, "core")
	old := releasepack.NewBundle("o/r", []releasepack.Package{{Project: "old", Ecosystem: "npm", Version: "1.0.0", Tag: "old@1.0.0", Files: []releasepack.File{{Name: "old-1.0.0.tgz", SHA256: sha("o")}}}})
	raw, _ := releasepack.BundleJSON(old)
	client, _ := githubrelease.FromEnvironment()
	repo := githubrelease.Repository{Owner: "o", Name: "r"}
	if _, err := githubrelease.Ensure(client, repo, githubrelease.Request{Tag: "core@0.5.2", Assets: []string{writeTemp(t, releasepack.BundleFile, raw)}}); err != nil {
		t.Fatal(err)
	}
	report := Report{Results: []ReportResult{{Item: npmPlan.Items[0], Packed: []releasepack.File{{Name: "core-0.5.2.tgz", SHA256: sha("c")}}}}}
	if err := attachBundle(s, &report, Options{Out: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}
	var b releasepack.Bundle
	if err := json.Unmarshal([]byte(gh.body("core@0.5.2", releasepack.BundleFile)), &b); err != nil || len(b.Packages) != 2 {
		t.Fatalf("the packages of the earlier run stay: %v %+v", err, b)
	}
	if !strings.Contains(gh.body("core@0.5.2", releasepack.BundleMarkdown), "old-1.0.0.tgz") {
		t.Fatal("INSTALL.md lists them too")
	}
}
