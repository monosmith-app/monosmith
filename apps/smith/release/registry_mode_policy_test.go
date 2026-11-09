package release

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/releasepack"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func projectWith(kind protocol.ProjectKind, reg *protocol.Registries, tags ...string) *projectgraph.Project {
	p := &projectgraph.Project{Project: protocol.Project{Name: "core", Kind: kind, Registries: reg}}
	p.Tags = tags
	return p
}

func TestTheRegistryModes(t *testing.T) {
	pack := &protocol.Registries{PackToRelease: true}
	none := &protocol.Registries{NoPublish: true}
	cases := []struct {
		name                  string
		p                     *projectgraph.Project
		noRegistry, packsFile bool
	}{
		{"a package with no setting", projectWith(protocol.ProjectPackage, nil), false, false},
		{"a package with a private registry", projectWith(protocol.ProjectPackage, &protocol.Registries{}), false, false},
		{"a package with registry = none", projectWith(protocol.ProjectPackage, none), true, false},
		{"a package with registry = github-release", projectWith(protocol.ProjectPackage, pack), true, true},
		{"an app with registry = github-release is never packed", projectWith(protocol.ProjectApp, pack), true, false},
		{"a native package is published by its jobs", projectWith(protocol.ProjectPackage, pack, NativeTag), true, false},
	}
	for _, c := range cases {
		if got := PublishesToNoRegistry(c.p); got != c.noRegistry {
			t.Errorf("%s: PublishesToNoRegistry = %v", c.name, got)
		}
		if got := PacksToRelease(c.p); got != c.packsFile {
			t.Errorf("%s: PacksToRelease = %v", c.name, got)
		}
	}
	if PackDir("core") != "dist/release/core/packages" {
		t.Fatal(PackDir("core"))
	}
}

func TestAPackedPackageIsInTheJournalAndTheReport(t *testing.T) {
	sum := strings.Repeat("ab", 32)
	res := ReportResult{Item: Item{Project: "core", Plugin: "node", Next: "0.5.2", Tag: "core@0.5.2"}, Status: "released", Pushed: true,
		GitHubRelease: "https://github.com/o/r/releases/tag/core@0.5.2", Packed: []releasepack.File{{Name: "core-0.5.2.tgz", SHA256: sum, Size: 3}}}
	var packs []string
	for _, e := range JournalEntries(res, time.Time{}) {
		if e.Step == "pack" {
			packs = append(packs, e.Target)
		}
	}
	if len(packs) != 1 || packs[0] != "core-0.5.2.tgz sha256:"+sum {
		t.Fatalf("the journal records the asset name and digest: %v", packs)
	}
	report := Report{Results: []ReportResult{res}}
	if !strings.Contains(report.Text(), "Attached to GitHub Release core@0.5.2: core-0.5.2.tgz") {
		t.Fatalf("text: %s", report.Text())
	}
	if !strings.Contains(report.Markdown(), "core-0.5.2.tgz") {
		t.Fatalf("markdown: %s", report.Markdown())
	}
}

func TestOnlyAPackageOrAppSentToARegistryPublishesToOne(t *testing.T) {
	npmApp := projectWith(protocol.ProjectApp, nil)
	npmApp.Release = &workspace.ProjectRelease{NPM: []string{"tool"}}
	for name, c := range map[string]struct {
		p    *projectgraph.Project
		want bool
	}{
		"package":       {projectWith(protocol.ProjectPackage, nil), true},
		"registry none": {projectWith(protocol.ProjectPackage, &protocol.Registries{NoPublish: true}), false},
		"app":           {projectWith(protocol.ProjectApp, nil), false},
		"app as npm":    {npmApp, true},
	} {
		if got := PublishesToRegistry(c.p); got != c.want {
			t.Errorf("%s: %v", name, got)
		}
	}
}
