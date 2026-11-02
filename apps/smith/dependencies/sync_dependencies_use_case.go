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

// SyncOptions say how Sync and Align run.
type SyncOptions struct {
	// Names limits the run to these dependencies; none means all of them.
	Names []string
	// Check only reports, and fails when declarations disagree.
	Check bool
	// DryRun settles and prints nothing: it returns the plan and writes nothing.
	DryRun bool
	// Install runs the install after the rewrite.
	Install bool
	// CrossMajor lets a declaration move to another major version.
	CrossMajor bool
	// Settle fixes the final target of each dependency (flags, questions) and
	// may refuse; nil takes smith's pick as it is.
	Settle func([]Drift) ([]Drift, error)
}

// Sync makes every project of each plugin declare a shared dependency at the
// same spec. With Check it only reports. After rewriting it installs, so
// lockfiles and the venv follow, unless Install is false.
func Sync(s *session.Session, o SyncOptions) (SyncReport, error) {
	return converge(s, o)
}

// Align is Sync for the named dependencies only, or for every fragmented one
// when none is named: `smith deps up --align`.
func Align(s *session.Session, o SyncOptions) (SyncReport, error) {
	return converge(s, o)
}

// converge brings the declarations of each plugin's projects to one spec in
// three steps: find what disagrees, let the person settle which version wins,
// write. Dependencies that [dependencies] allow lists are left alone; with
// names, only those are looked at.
func converge(s *session.Session, o SyncOptions) (SyncReport, error) {
	var report SyncReport
	type found struct {
		client   *host.Client
		projects []protocol.Project
		drift    []Drift
	}
	var all []found
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
			return slices.Contains(s.Config.Dependencies.Allow, d.Name) || (len(o.Names) > 0 && !slices.Contains(o.Names, d.Name))
		})
		drift := FindDrift(name, declared, local)
		if o.CrossMajor {
			for i := range drift {
				drift[i].Fixable = append(drift[i].Fixable, drift[i].Crossing...)
				drift[i].Crossing = nil
			}
		}
		all = append(all, found{c, projects, drift})
	}
	var flat []Drift
	for _, f := range all {
		flat = append(flat, f.drift...)
	}
	if !o.Check && len(flat) > 0 && o.Settle != nil {
		settled, err := o.Settle(flat)
		report.Drift = settled
		if err != nil {
			return report, err
		}
		flat = settled
	} else {
		report.Drift = flat
	}
	if o.Check && len(flat) > 0 {
		return report, ErrDrift
	}
	if o.Check || o.DryRun {
		return report, nil
	}
	next := 0
	for _, f := range all {
		var changes []protocol.Rewrite
		for _, d := range flat[next : next+len(f.drift)] {
			if o.CrossMajor {
				d.Fixable = append(d.Fixable, d.Crossing...)
			}
			for _, site := range d.Fixable {
				changes = append(changes, protocol.Rewrite{Declaration: site, Spec: d.Target})
			}
		}
		next += len(f.drift)
		if len(changes) == 0 {
			continue
		}
		files, err := rewriteAndInstall(f.client, changes, f.projects, o.Install)
		report.Files = append(report.Files, files...)
		if err != nil {
			return report, fmt.Errorf("plugin %s: %w", f.client.Info.Plugin.Name, err)
		}
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
		for _, site := range d.Crossing {
			fmt.Fprintf(w, "  %-12s %s left alone: it would move from major %d to %d, which can break the project. Pass --yes to allow it (%s)\n", where(site), site.Spec, majorOf(site.Spec), majorOf(d.Target), site.File)
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
	if reason == "chosen" {
		return "(your choice)"
	}
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
