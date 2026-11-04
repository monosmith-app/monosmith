package dependencies

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// uninstallAtRoot removes packages from the manifest at the workspace root. The
// plugin that declares them there (the one whose Declared lists them with no
// project) removes them, from whichever section holds them, a development
// dependency included.
//
// Parameters: s is the open session; packages are the names to remove; log
// receives the report.
// Returns: an error when no plugin declares the packages at the root, when only
// some of them are declared there, or when the plugin fails.
func uninstallAtRoot(s *session.Session, packages []string, log io.Writer) error {
	for _, c := range s.Plugins.Clients {
		name := c.Info.Plugin.Name
		if !c.Supports(protocol.MethodDeclared) || !c.Supports(protocol.MethodUninstall) {
			continue
		}
		var projects []protocol.Project
		for _, p := range s.Graph.Projects {
			if p.Plugin == name {
				projects = append(projects, p.Project)
			}
		}
		var declared protocol.DeclaredResult
		if err := c.Call(protocol.MethodDeclared, protocol.DeclaredParams{Projects: projects}, &declared, timeout); err != nil {
			continue
		}
		missing, have := undeclared(declared.Declarations, "", packages)
		if len(have) == 0 {
			continue
		}
		if len(missing) > 0 {
			return fmt.Errorf("the workspace root does not declare %s. It declares: %s", strings.Join(missing, ", "), strings.Join(have, ", "))
		}
		var res protocol.DependencyResult
		if err := c.Call(protocol.MethodUninstall, protocol.UninstallParams{Project: protocol.Project{Root: RootTarget}, Packages: packages}, &res, timeout); err != nil {
			return fmt.Errorf("the workspace root (plugin %s): %w", name, err)
		}
		report(log, "(root): removed "+strings.Join(packages, ", "), res)
		return nil
	}
	return errors.New("the workspace root does not declare " + strings.Join(packages, ", ") + ": no plugin lists a package there")
}
