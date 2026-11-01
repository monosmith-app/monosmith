package conformance

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// declarer is the good language that also describes declared projects: it
// describes a declared folder that exists and refuses one that does not.
type declarer struct{ good }

func (d *declarer) DescribesDeclared() {}

func (d *declarer) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	res, err := d.good.Discover(protocol.DiscoverParams{Manifests: p.Manifests})
	for _, decl := range p.Declared {
		return res, fmt.Errorf("%s: the folder does not exist", decl.Root)
	}
	return res, err
}

// inventor describes a declared project whatever the folder, which is wrong.
type inventor struct{ declarer }

func (i *inventor) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	res, _ := i.good.Discover(protocol.DiscoverParams{Manifests: p.Manifests})
	for _, decl := range p.Declared {
		res.Projects = append(res.Projects, protocol.Project{Name: decl.Name, Root: decl.Root, Kind: protocol.ProjectLib})
	}
	return res, nil
}

func failedChecks(r Report) string {
	var out []string
	for _, res := range r.Results {
		if !res.OK {
			out = append(out, res.Check)
		}
	}
	return strings.Join(out, "|")
}

func TestAPluginThatDescribesDeclaredProjectsRefusesAFolderThatIsNotThere(t *testing.T) {
	ws := t.TempDir()
	r := Check(connect(t, &declarer{}, ws), Options{Workspace: ws})
	if failed := failedChecks(r); failed != "" {
		t.Fatalf("failures: %s", failed)
	}
	if !strings.Contains(fmt.Sprint(r.Results), "a declared folder that does not exist is refused, not invented") {
		t.Errorf("the check did not run: %v", r.Results)
	}
	bad := Check(connect(t, &inventor{}, t.TempDir()), Options{})
	if !strings.Contains(failedChecks(bad), "a declared folder that does not exist is refused, not invented") {
		t.Errorf("a plugin that invents a project for a missing folder must fail: %s", failedChecks(bad))
	}
}

func TestAPluginThatDescribesNoDeclaredProjectIgnoresADeclaration(t *testing.T) {
	ws := t.TempDir()
	r := Check(connect(t, &good{}, ws), Options{Workspace: ws})
	if !strings.Contains(fmt.Sprint(r.Results), "a plugin that describes no declared project ignores a declaration") {
		t.Errorf("the check did not run: %v", r.Results)
	}
}
