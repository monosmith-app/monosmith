package dependencies

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ErrDrift is sync --check finding declarations that disagree.
var ErrDrift = errors.New("projects declare shared dependencies at different versions; run smith sync")

// SyncReport is `smith sync --json`.
type SyncReport struct {
	Drift []Drift `json:"drift"`
	// Files are the manifests sync rewrote.
	Files []string `json:"files,omitempty"`
	// Unchecked are plugins whose projects sync cannot compare.
	Unchecked []string `json:"unchecked,omitempty"`
}

// Sync makes every project of each plugin declare a shared dependency at the
// same spec. With check it only reports. After rewriting it installs, so
// lockfiles and the venv follow, unless install is false.
func Sync(s *session.Session, check, install bool) (SyncReport, error) {
	return converge(s, nil, check, install)
}

// Align is Sync for the named dependencies only, or for every fragmented one
// when none is named: `smith deps up --align`.
func Align(s *session.Session, names []string, check, install bool) (SyncReport, error) {
	return converge(s, names, check, install)
}

// converge brings the declarations of each plugin's projects to one spec.
// Dependencies that [dependencies] allow lists are left alone; with names,
// only those are looked at.
func converge(s *session.Session, names []string, check, install bool) (SyncReport, error) {
	var report SyncReport
	for _, c := range s.Plugins.Clients {
		name := c.Info.Plugin.Name
		projects, local := servedBy(s, name)
		if len(projects) == 0 {
			continue
		}
		if !c.Supports(protocol.MethodDeclared) {
			if c.Supports(protocol.MethodInstall) {
				report.Unchecked = append(report.Unchecked, name)
			}
			continue
		}
		var declared protocol.DeclaredResult
		if err := c.Call(protocol.MethodDeclared, protocol.DeclaredParams{Projects: projects}, &declared, timeout); err != nil {
			return report, fmt.Errorf("plugin %s: %w", name, err)
		}
		declared.Declarations = slices.DeleteFunc(declared.Declarations, func(d protocol.Declaration) bool {
			return slices.Contains(s.Config.Dependencies.Allow, d.Name) || (len(names) > 0 && !slices.Contains(names, d.Name))
		})
		drift := FindDrift(name, declared, local)
		report.Drift = append(report.Drift, drift...)
		if check {
			continue
		}
		var changes []protocol.Rewrite
		for _, d := range drift {
			for _, site := range d.Fixable {
				changes = append(changes, protocol.Rewrite{Declaration: site, Spec: d.Target})
			}
		}
		if len(changes) == 0 {
			continue
		}
		files, err := rewriteAndInstall(c, changes, projects, install)
		report.Files = append(report.Files, files...)
		if err != nil {
			return report, fmt.Errorf("plugin %s: %w", name, err)
		}
	}
	if check && len(report.Drift) > 0 {
		return report, ErrDrift
	}
	return report, nil
}

// servedBy is the projects a plugin serves, and the names they publish
// under, which sync leaves alone: the workspace builds those itself.
func servedBy(s *session.Session, plugin string) ([]protocol.Project, map[string]bool) {
	var projects []protocol.Project
	local := map[string]bool{}
	for _, p := range s.Graph.Projects {
		local[p.Name] = true
		for _, e := range p.Exports {
			local[e] = true
		}
		if p.Plugin == plugin {
			projects = append(projects, p.Project)
		}
	}
	return projects, local
}

func rewriteAndInstall(c *host.Client, changes []protocol.Rewrite, projects []protocol.Project, install bool) ([]string, error) {
	var res protocol.DependencyResult
	if err := c.Call(protocol.MethodRewrite, protocol.RewriteParams{Changes: changes}, &res, timeout); err != nil {
		return nil, err
	}
	if install && c.Supports(protocol.MethodInstall) {
		var installed protocol.DependencyResult
		if err := c.Call(protocol.MethodInstall, protocol.InstallParams{Projects: projects}, &installed, timeout); err != nil {
			return res.Files, err
		}
		res.Files = append(res.Files, installed.Files...)
	}
	return res.Files, nil
}

// PrintSync writes the report for people.
func PrintSync(w io.Writer, r SyncReport, check bool) {
	for _, d := range r.Drift {
		fmt.Fprintf(w, "%s (%s) → %s  %s\n", d.Name, d.Plugin, d.Target, hint(d.Reason))
		for _, site := range d.Fixable {
			verb := "rewritten"
			if check {
				verb = "would be rewritten"
			}
			fmt.Fprintf(w, "  %-12s %s %s (%s)\n", where(site), site.Spec, verb, site.File)
		}
		for _, site := range d.Blocked {
			fmt.Fprintf(w, "  %-12s %s left alone: smith does not rewrite this kind of spec (%s)\n", where(site), site.Spec, site.File)
		}
	}
	if len(r.Drift) == 0 {
		fmt.Fprintln(w, "✔ Every project agrees on every shared dependency version.")
	} else if !check {
		fmt.Fprintf(w, "✔ %d dependencies converged (%s)\n", len(r.Drift), strings.Join(r.Files, ", "))
	}
	for _, p := range r.Unchecked {
		fmt.Fprintf(w, "⊘ %s: not compared; its plugin does not list declarations\n", p)
	}
	for _, d := range r.Drift {
		if d.Plugin == "go" {
			fmt.Fprintln(w, "ℹ A go.work already builds every module with the highest version any of them requires; aligning go.mod keeps each module building the same on its own.")
			break
		}
	}
}

func hint(reason string) string {
	if reason == "installed" {
		return "(the version installed)"
	}
	return "(the highest declared)"
}

func where(d protocol.Declaration) string {
	if d.Project == "" {
		return "(root)"
	}
	return d.Project
}
