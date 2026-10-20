package dependencies

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Checks that smith doctor reports under the one-version policy.
const (
	CheckOneVersion = "every production dependency has one version across the workspace"
	CheckDevAtRoot  = "dev tools are declared once, at the root"
)

// CheckOneVersionPolicy applies [dependencies] one_version: it asks every
// plugin that lists declarations and reports the production dependencies its
// projects declare at different versions, and the dev tools declared outside
// the root. Without the policy it reports nothing.
func CheckOneVersionPolicy(s *session.Session) []protocol.Finding {
	if !s.Config.Dependencies.OneVersion {
		return nil
	}
	allow := s.Config.Dependencies.Allow
	var fragmented, misplaced []string
	for _, c := range s.Plugins.Clients {
		name := c.Info.Plugin.Name
		projects, local := servedBy(s, name)
		if len(projects) == 0 || !c.Supports(protocol.MethodDeclared) {
			continue
		}
		var declared protocol.DeclaredResult
		if err := c.Call(protocol.MethodDeclared, protocol.DeclaredParams{Projects: projects}, &declared, timeout); err != nil {
			return []protocol.Finding{{Check: "plugin " + name + " lists its declared dependencies", Detail: err.Error(), Remedy: "report this to the plugin's maintainers"}}
		}
		for _, d := range FindFragmentation(name, declared, local, allow) {
			fragmented = append(fragmented, describe(d))
		}
		for _, m := range FindMisplacedDevTools(name, declared, local, allow) {
			misplaced = append(misplaced, fmt.Sprintf("%s (%s) is declared in %s", m.Name, name, strings.Join(m.Sites, ", ")))
		}
	}
	return []protocol.Finding{
		{
			Check: CheckOneVersion, OK: len(fragmented) == 0,
			Detail: strings.Join(fragmented, "; "),
			Remedy: "`smith deps up --align` brings each to one version; a dependency that may differ is listed in smith.toml [dependencies] allow = [...]",
		},
		{
			Check: CheckDevAtRoot, OK: len(misplaced) == 0, Warning: true,
			Detail: strings.Join(misplaced, "; "),
			Remedy: "declare the dev tool once, in the root manifest, and remove it from the projects (`smith uninstall -w <project> <package>`)",
		},
	}
}

// describe names a fragmented dependency and every spec declared for it.
func describe(d Drift) string {
	sites := append(append([]protocol.Declaration{}, d.Fixable...), d.Blocked...)
	var parts []string
	for _, s := range sites {
		parts = append(parts, fmt.Sprintf("%s %s", where(s), s.Spec))
	}
	return fmt.Sprintf("%s (%s) should be %s everywhere, but is %s", d.Name, d.Plugin, d.Target, strings.Join(parts, ", "))
}
