package kindchange

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func subject(from, to protocol.ProjectKind) Subject {
	return Subject{Name: "core", Language: "rust", From: from, To: to, Root: "libs/core", NewRoot: "packages/core"}
}

func reasons(rs []Refusal) string {
	var out []string
	for _, r := range rs {
		out = append(out, r.Reason)
	}
	return strings.Join(out, "\n")
}

func TestSameKindIsRefusedAndCannotBeForced(t *testing.T) {
	refusals, _ := Review(subject(protocol.ProjectLib, protocol.ProjectLib))
	if len(refusals) != 1 || refusals[0].Forceable || !strings.Contains(refusals[0].Reason, "already a lib") {
		t.Fatalf("%+v", refusals)
	}
}

func TestAnExistingTargetFolderCannotBeForced(t *testing.T) {
	s := subject(protocol.ProjectLib, protocol.ProjectPackage)
	s.TargetExists = true
	refusals, _ := Review(s)
	if len(refusals) != 1 || refusals[0].Forceable || !strings.Contains(refusals[0].Reason, "packages/core already exists") {
		t.Fatalf("%+v", refusals)
	}
}

func TestAnAppThatOthersImportIsRefusedButMayBeForced(t *testing.T) {
	s := subject(protocol.ProjectLib, protocol.ProjectApp)
	s.Dependents = []Related{{Name: "cli", Kind: protocol.ProjectApp, Language: "rust"}}
	refusals, _ := Review(s)
	if len(refusals) != 1 || !refusals[0].Forceable || !strings.Contains(refusals[0].Reason, "cli imports it") {
		t.Fatalf("%+v", refusals)
	}
	s.Dependents = nil
	if refusals, _ := Review(s); len(refusals) != 0 {
		t.Fatalf("nothing imports it: %+v", refusals)
	}
}

func TestAPackageThatPublishedPackagesImportStaysPublished(t *testing.T) {
	s := subject(protocol.ProjectPackage, protocol.ProjectLib)
	s.Dependents = []Related{{Name: "sdk", Kind: protocol.ProjectPackage, Language: "rust"}, {Name: "web", Kind: protocol.ProjectApp, Language: "rust"}}
	refusals, _ := Review(s)
	got := reasons(refusals)
	if !strings.Contains(got, "sdk is a published package that imports core") || strings.Contains(got, "web") {
		t.Fatalf("only the published dependents count:\n%s", got)
	}
}

func TestAReleasedPackageIsNotQuietlyOrphaned(t *testing.T) {
	s := subject(protocol.ProjectPackage, protocol.ProjectLib)
	s.LatestTag = "core@1.2.0"
	refusals, _ := Review(s)
	if len(refusals) != 1 || !refusals[0].Forceable || !strings.Contains(refusals[0].Reason, "core@1.2.0") {
		t.Fatalf("%+v", refusals)
	}
	// Becoming a package, or moving between app and lib, orphans nothing.
	s.From, s.To = protocol.ProjectLib, protocol.ProjectPackage
	if refusals, _ := Review(s); len(refusals) != 0 {
		t.Fatalf("%+v", refusals)
	}
}

func TestWarningsNameWhatTheChangeLeavesBehind(t *testing.T) {
	s := subject(protocol.ProjectApp, protocol.ProjectPackage)
	s.HasDockerfile, s.HasReleaseSettings = true, true
	s.Dependencies = []Related{{Name: "util", Kind: protocol.ProjectLib, Language: "rust"}}
	s.Dependents = []Related{{Name: "viz", Kind: protocol.ProjectLib, Language: "python"}}
	_, warnings := Review(s)
	joined := strings.Join(warnings, "\n")
	for _, want := range []string{"imports util, which is not published", "will publish core", "Dockerfile", "[projects.core.release]", "viz depends on core through its schemas"} {
		if !strings.Contains(joined, want) {
			t.Errorf("no warning says %q:\n%s", want, joined)
		}
	}
}
