package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func pkg(name, plugin string, reg *protocol.Registries) *projectgraph.Project {
	p := &projectgraph.Project{Project: protocol.Project{Name: name, Kind: protocol.ProjectPackage, Registries: reg}}
	p.Plugin = plugin
	return p
}

func TestAWorkspaceWithoutTheModeHasNoFindings(t *testing.T) {
	got := githubReleaseFindings([]*projectgraph.Project{pkg("core", "node", nil)}, true, func(string) bool { return true }, func(string) string { return "" })
	if len(got) != 0 {
		t.Fatalf("%+v", got)
	}
}

func TestTheModeIsReportedAndAPublicPackageWithoutItsCredentialWarns(t *testing.T) {
	pack := &protocol.Registries{PackToRelease: true}
	projects := []*projectgraph.Project{pkg("core", "node", pack), pkg("lib", "python", nil), pkg("go", "go", nil), pkg("tagged", "ruby", &protocol.Registries{NoPublish: true})}
	got := githubReleaseFindings(projects, true, func(string) bool { return true }, func(string) string { return "" })
	var warnings []string
	listed := false
	for _, f := range got {
		if f.Warning {
			warnings = append(warnings, f.Check+" "+f.Detail+" "+f.Remedy)
		}
		if strings.Contains(f.Check, "1 package(s) attach to GitHub Releases") && f.OK {
			listed = true
		}
		if !f.OK && !f.Warning {
			t.Errorf("nothing fails here: %+v", f)
		}
	}
	if !listed {
		t.Errorf("the packages that attach are listed: %+v", got)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "PYPI_TOKEN") || !strings.Contains(warnings[0], `registry = "github-release"`) {
		t.Fatalf("only lib asks for a public credential: %v", warnings)
	}
	set := githubReleaseFindings(projects, true, func(string) bool { return true }, func(env string) string { return "x" })
	for _, f := range set {
		if f.Warning {
			t.Fatalf("a credential that is set is no warning: %+v", f)
		}
	}
}

func TestTheModeNeedsGitHubReleasesAndAPackingPlugin(t *testing.T) {
	pack := &protocol.Registries{PackToRelease: true}
	got := githubReleaseFindings([]*projectgraph.Project{pkg("core", "node", pack)}, false, func(string) bool { return false }, func(string) string { return "" })
	failed := 0
	for _, f := range got {
		if !f.OK && !f.Warning {
			failed++
		}
	}
	if failed != 2 {
		t.Fatalf("github_releases off and a plugin that cannot pack are two failures: %+v", got)
	}
}
