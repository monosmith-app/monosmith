package dependencies

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// checkRootInstall says whether packages can be added at the workspace root.
//
// Parameters: s is the open session.
// Returns: an error naming the supported root when the workspace has none.
func checkRootInstall(s *session.Session) error {
	if c := s.Plugins.Get(RootPlugin); c != nil && c.Supports(protocol.MethodInstall) {
		return nil
	}
	return errors.New("this workspace has no root manifest smith can add packages to. Only the Node root package.json is supported. Use smith install -w <project> <package>")
}

// installAtRoot adds packages to the manifest at the workspace root, shared by
// every project, with the node plugin and no --workspace.
//
// Parameters: s is the open session; packages are the names; dev adds them as
// development dependencies; log receives the report.
// Returns: an error when the root cannot take packages, or when the plugin
// fails (the manifest and lock file are put back).
func installAtRoot(s *session.Session, packages []string, dev bool, log io.Writer) error {
	if err := checkRootInstall(s); err != nil {
		return err
	}
	var res protocol.DependencyResult
	root := protocol.Project{Root: RootTarget}
	err := guarded([]string{s.Root}, func() error {
		if err := s.Plugins.Get(RootPlugin).Call(protocol.MethodInstall, protocol.InstallParams{Project: &root, Packages: packages, Dev: dev}, &res, timeout); err != nil {
			return fmt.Errorf("the workspace root (plugin %s): %w", RootPlugin, err)
		}
		return nil
	})
	if err != nil {
		if hint := UnknownNameHint(s.Graph.Names(), packages); hint != "" {
			err = fmt.Errorf("%w\n%s", err, hint)
		}
		return err
	}
	report(log, "(root): added "+strings.Join(packages, ", "), res)
	return nil
}
