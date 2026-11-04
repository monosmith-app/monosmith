package dependencies

import (
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/internallink"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// PlanInstall prints what smith install would do, and changes nothing: for each
// target project the packages, the section they go into, and the newest release
// the registry reports for a name typed with no version. A name that is a library
// of the workspace is shown as a link. The registry description of a package is
// not part of the plugin protocol, so it is not shown.
//
// Parameters: s is the open session; targets are the -w projects; packages the
// names; dev is -D; registry is --registry; out receives the plan.
// Returns: an error when a target does not resolve or no target is given for a package.
func PlanInstall(s *session.Session, targets, packages []string, dev, registry bool, out io.Writer) error {
	if len(packages) == 0 {
		fmt.Fprintln(out, "Installs what every project already declares, with its own toolchain.")
		fmt.Fprintln(out, "Dry run: nothing was changed.")
		return nil
	}
	if len(targets) == 0 {
		return ErrNoTarget
	}
	projects, err := resolve(s.Graph, targets)
	if err != nil {
		return err
	}
	section := "its dependencies"
	if dev {
		section = "its development dependencies"
	}
	for _, p := range projects {
		fmt.Fprintf(out, "%s (%s) gets:\n", p.Name, p.Plugin)
		var bare []string
		for _, name := range packages {
			if _, ok := splitVersion(name); ok {
				bare = append(bare, name)
			}
		}
		var latest protocol.LatestResult
		if len(bare) > 0 {
			// A registry that cannot be asked leaves the versions out; the install itself still works.
			_, _ = s.Plugins.Call(p.Plugin, protocol.MethodLatest, protocol.LatestParams{Names: bare}, &latest, timeout)
		}
		for _, name := range packages {
			stripped, _ := internallink.StripPrefix(name)
			if !registry && internallink.Match(s.Graph, stripped) != nil {
				fmt.Fprintf(out, "  %s: a library of this workspace, linked; nothing is downloaded\n", name)
				continue
			}
			fmt.Fprintf(out, "  %s: %s\n", name, versionNote(name, latest))
		}
		fmt.Fprintf(out, "  written to %s\n", section)
	}
	fmt.Fprintln(out, "Dry run: nothing was changed.")
	return nil
}

// splitVersion returns the name of a package typed with no version, and false
// for one that carries a version spec (name@1.2, name>=2, name==1) or a prefix.
func splitVersion(name string) (string, bool) {
	if strings.HasPrefix(name, internallink.WorkspacePrefix) {
		return "", false
	}
	if i := strings.LastIndex(name, "@"); i > 0 {
		return "", false
	}
	if strings.ContainsAny(name, "=<>~!") {
		return "", false
	}
	return name, true
}

// versionNote says what version a package is added at.
func versionNote(name string, latest protocol.LatestResult) string {
	if _, ok := splitVersion(name); !ok {
		return "the version you typed"
	}
	if v := latest.Latest[name]; v != "" {
		return "latest release " + v
	}
	if reason := latest.Unreachable[name]; reason != "" {
		return "the registry could not be asked (" + reason + ")"
	}
	return "not found in the registry, so the install will fail: check the name"
}
