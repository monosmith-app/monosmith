package electron

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Doctor checks that no project of the electron stack smith made holds a copy of a smith
// script (scripts/assemble.mjs, start.mjs, package.mjs). The app smith makes
// runs `smith electron assemble` and Electron's own commands instead.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	findings := []protocol.Finding{}
	for _, p := range params.Projects {
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
		if _, err := os.Stat(filepath.Join(dir, "scripts", "assemble.mjs")); err != nil {
			continue
		}
		if f, found := functionhost.CopiedScriptsFinding(workspaceRoot, p, functionhost.Electron); found {
			findings = append(findings, f)
		}
	}
	return protocol.DoctorResult{Findings: findings}
}

// Doctor implements serve.DoctorContributor.
func (p *Electron) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return Doctor(p.root, params), nil
}
