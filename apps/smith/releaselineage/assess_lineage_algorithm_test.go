package releaselineage

import (
	"strings"
	"testing"
)

func TestAssess(t *testing.T) {
	tags := []string{
		"sdk@0.1.0", "sdk@0.2.0", "sdk@0.3.0-rc.1",
		"@acme/web@1.4.0", "web@1.2.0",
		"ui@2.0.0", "@acme/ui@1.0.0",
		"api@1.0.0",
		"packages/gokit/v0.3.0",
	}
	subjects := []Subject{
		// Tags only under the scoped package name: stranded.
		{Project: "web", Prefix: "web@", Alternatives: []string{"@acme/web@"}, Tags: tags, Published: "1.4.0"},
		// Its own tags are ahead of the old name: fine.
		{Project: "ui", Prefix: "ui@", Alternatives: []string{"@acme/ui@"}, Tags: tags, Published: "2.0.0"},
		// Tagged 0.2.0, registry has 0.4.0: behind.
		{Project: "sdk", Prefix: "sdk@", Tags: tags, Published: "0.4.0"},
		// Published but never tagged.
		{Project: "kit", Prefix: "kit@", Tags: tags, Published: "3.1.0"},
		// Not published yet: a first release, fine.
		{Project: "new", Prefix: "new@", Tags: tags},
		// Equal to the registry: fine.
		{Project: "api", Prefix: "api@", Tags: tags, Published: "1.0.0"},
		// Go: no registry answer, only its own tags.
		{Project: "gokit", Prefix: "packages/gokit/v", Tags: tags},
		// The registry could not be asked: a warning.
		{Project: "py", Prefix: "py@", Tags: tags, Unreachable: true, Registry: "https://pypi.org/simple/", Detail: "connection refused"},
	}
	got := map[string]Finding{}
	for _, f := range Assess(subjects) {
		got[f.Project] = f
	}
	if len(got) != 4 {
		t.Fatalf("findings for %v", got)
	}
	for project, want := range map[string]struct {
		blocking bool
		text     string
	}{
		"web": {true, `newest release tag is @acme/web@1.4.0, under another name than its prefix web@, so the next release would start from web@1.2.0.`},
		"sdk": {true, "tagged sdk@0.2.0 but the registry already has 0.4.0"},
		"kit": {true, "published at 3.1.0 according to the registry, but has no release tag"},
		"py":  {false, "could not ask https://pypi.org/simple/ about py (connection refused)"},
	} {
		f := got[project]
		if f.Blocking != want.blocking || !strings.Contains(f.Message, want.text) {
			t.Errorf("%s: %+v\nwant %q", project, f, want.text)
		}
	}
	if remedy := strings.Join(got["web"].Remedy, "\n"); !strings.Contains(remedy, "git tag web@1.4.0 @acme/web@1.4.0^{commit}") || !strings.Contains(remedy, `tag_prefix = "@acme/web@"`) {
		t.Errorf("web remedy: %s", remedy)
	}
	if len(got["kit"].Remedy) < 2 || strings.Contains(got["kit"].Sentence(), "see the") {
		t.Errorf("kit remedy must be real steps: %v", got["kit"].Remedy)
	}
}

func TestTwoProjectsMaySharePackageNameInDifferentRegistries(t *testing.T) {
	// A Python and a Dart package, both called opencraw, in one repository.
	tags := []string{"opencraw@0.3.0", "opencraw-dart@0.1.0"}
	subjects := []Subject{
		{Project: "opencraw", Prefix: "opencraw@", Tags: tags},
		// The Dart package's own name makes opencraw@ an earlier name, but
		// another project releases under it.
		{Project: "opencraw-dart", Prefix: "opencraw-dart@", Alternatives: []string{"opencraw@"}, Tags: tags},
	}
	if got := Assess(subjects); len(got) != 0 {
		t.Errorf("findings %+v; a tag another project's prefix owns is not stranded", got)
	}
}

func TestAnExplicitPrefixWithNoTagsYetIsAFirstRelease(t *testing.T) {
	tags := []string{"@acme/web@1.4.0"}
	stranded := Subject{Project: "web", Prefix: "web@", Alternatives: []string{"@acme/web@"}, Tags: tags}
	if got := Assess([]Subject{stranded}); len(got) != 1 || !got[0].Blocking {
		t.Fatalf("an implicit prefix with only old-name tags is stranded: %+v", got)
	}
	stranded.Explicit = true
	if got := Assess([]Subject{stranded}); len(got) != 0 {
		t.Errorf("findings %+v; a deliberate prefix with no tags is trusted", got)
	}
}
