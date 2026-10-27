package pipeline

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/declaredimports"
	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// shownImports caps how many undeclared imports a finding lists.
const shownImports = 5

// CheckPackageImports is doctor's view of a package's deps-check task: every
// package it imports is declared where installing it from npm installs it.
func CheckPackageImports(workspaceRoot string, p protocol.Project) protocol.Finding {
	f := protocol.Finding{
		Check:  p.Root + " declares the packages it imports",
		Remedy: "add each to dependencies in " + p.Root + "/package.json (peerDependencies when the installing app provides it), then run `npm install`",
	}
	undeclared, err := declaredimports.Check(workspaceRoot, p.Root)
	if err != nil {
		f.Detail = err.Error()
		f.Remedy = "smith could not read " + p.Root + " to check its imports: fix what the problem says, then check again"
		return f
	}
	f.OK = len(undeclared) == 0
	lines := make([]string, 0, shownImports+1)
	for i, u := range undeclared {
		if i == shownImports {
			lines = append(lines, fmt.Sprintf("and %d more (smith run deps-check -p %s)", len(undeclared)-shownImports, p.Name))
			break
		}
		lines = append(lines, u.String())
	}
	f.Detail = strings.Join(lines, "; ")
	if !f.OK {
		f.Fix = manifestfixes.DeclareImportsFix(workspaceRoot, p, undeclared)
	}
	return f
}
