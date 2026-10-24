package manifestfixes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/declaredimports"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// InstallMissingPackagesFix installs the packages a root config imports and
// node_modules lacks. A package the root package.json already declares only
// needs `npm install`; one it does not is added with `npm install -D`, at the
// range the workspace uses elsewhere when it does. It is not safe: installing
// changes package.json and the lockfile.
func InstallMissingPackagesFix(workspaceRoot string, missing []string) *protocol.Fix {
	if len(missing) == 0 {
		return nil
	}
	declared := rootDeclarations(workspaceRoot)
	known := declaredimports.WorkspaceSpecs(workspaceRoot)
	var undeclared []string
	needInstall := false
	for _, name := range missing {
		if declared[name] {
			needInstall = true
			continue
		}
		pkg := name
		if specs := known[name]; len(specs) > 0 {
			pkg += "@" + specs[0].Spec
		}
		undeclared = append(undeclared, pkg)
	}
	sort.Strings(undeclared)
	fix := &protocol.Fix{Description: "install the packages the ESLint config imports"}
	if needInstall {
		fix.Steps = append(fix.Steps, protocol.FixStep{Description: "install what package.json declares", Argv: []string{"npm", "install", "--no-audit", "--no-fund"}})
	}
	if len(undeclared) > 0 {
		argv := append([]string{"npm", "install", "-D", "--no-audit", "--no-fund"}, undeclared...)
		fix.Steps = append(fix.Steps, protocol.FixStep{Description: "add " + strings.Join(undeclared, ", ") + " to the root devDependencies", Argv: argv})
	}
	return fix
}

// rootDeclarations are the packages the root package.json declares, in any section.
func rootDeclarations(workspaceRoot string) map[string]bool {
	out := map[string]bool{}
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, "package.json"))
	if err != nil {
		return out
	}
	var pkg struct {
		Dependencies, DevDependencies, OptionalDependencies, PeerDependencies map[string]string
	}
	if json.Unmarshal(raw, &pkg) != nil {
		return out
	}
	for _, section := range []map[string]string{pkg.Dependencies, pkg.DevDependencies, pkg.OptionalDependencies, pkg.PeerDependencies} {
		for name := range section {
			out[name] = true
		}
	}
	return out
}
