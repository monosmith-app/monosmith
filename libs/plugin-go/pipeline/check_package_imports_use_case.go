package pipeline

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-go/declaredimports"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// shownImports caps how many undeclared imports a finding lists.
const shownImports = 5

// CheckPackageImports is doctor's view of a package's deps-check task: every
// module it imports is required by its go.mod.
func CheckPackageImports(workspaceRoot string, p protocol.Project) protocol.Finding {
	f := protocol.Finding{
		Check:   p.Root + " requires the modules it imports",
		Project: p.Name,
		Remedy:  "run `go mod tidy` in " + p.Root + " to add what its go.mod lacks",
	}
	undeclared, err := declaredimports.Check(workspaceRoot, p.Root)
	if err != nil {
		f.Problem, f.Detail = "smith could not read its files to check its imports", err.Error()
		f.Remedy = "smith could not read " + p.Root + " to check its imports: fix what the problem says, then check again"
		return f
	}
	f.OK = len(undeclared) == 0
	if !f.OK {
		f.Problem = fmt.Sprintf("imports %d module(s) its go.mod does not require", len(undeclared))
	}
	lines := make([]string, 0, shownImports+1)
	for i, u := range undeclared {
		if i == shownImports {
			lines = append(lines, fmt.Sprintf("and %d more (smith run deps-check -p %s)", len(undeclared)-shownImports, p.Name))
			break
		}
		lines = append(lines, u.String())
	}
	f.Detail = strings.Join(lines, "; ")
	return f
}
