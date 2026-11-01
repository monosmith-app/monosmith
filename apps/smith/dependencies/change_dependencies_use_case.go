package dependencies

import (
	"errors"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/nearmatch"
	"io"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const timeout = 20 * time.Minute

// ErrNoTarget is adding packages without saying to which project.
var ErrNoTarget = errors.New("adding a package needs a target project: smith install -w <project> <package>. A bare smith install (no package) installs the whole workspace")

// Install adds packages to each target project, or, with no target and no
// package, installs every project of every plugin that can.
func Install(s *session.Session, targets, packages []string, dev bool, log io.Writer) error {
	if len(targets) == 0 {
		if len(packages) > 0 {
			return ErrNoTarget
		}
		return installAll(s, log)
	}
	projects, err := resolve(s.Graph, targets)
	if err != nil {
		return err
	}
	for _, p := range projects {
		var res protocol.DependencyResult
		ok, err := s.Plugins.Call(p.Plugin, protocol.MethodInstall, protocol.InstallParams{Project: &p.Project, Packages: packages, Dev: dev}, &res, timeout)
		if err := outcome(p, ok, err); err != nil {
			return err
		}
		what := "installed"
		if len(packages) > 0 {
			what = "added " + strings.Join(packages, ", ")
		}
		report(log, p.Name+": "+what, res)
	}
	return nil
}

// Uninstall removes packages from each target project.
func Uninstall(s *session.Session, targets, packages []string, log io.Writer) error {
	if len(targets) == 0 {
		return errors.New("a dependency belongs to the project that uses it: smith uninstall -w <project> <package>")
	}
	projects, err := resolve(s.Graph, targets)
	if err != nil {
		return err
	}
	for _, p := range projects {
		if c := s.Plugins.Get(p.Plugin); c != nil && c.Supports(protocol.MethodDeclared) {
			var declared protocol.DeclaredResult
			if err := c.Call(protocol.MethodDeclared, protocol.DeclaredParams{Projects: []protocol.Project{p.Project}}, &declared, timeout); err == nil {
				if missing, have := undeclared(declared.Declarations, p.Name, packages); len(missing) > 0 {
					return fmt.Errorf("%s does not declare %s. It declares: %s", p.Name, strings.Join(missing, ", "), strings.Join(have, ", "))
				}
			}
		}
		var res protocol.DependencyResult
		ok, err := s.Plugins.Call(p.Plugin, protocol.MethodUninstall, protocol.UninstallParams{Project: p.Project, Packages: packages}, &res, timeout)
		if err := outcome(p, ok, err); err != nil {
			return err
		}
		report(log, p.Name+": removed "+strings.Join(packages, ", "), res)
	}
	return nil
}

func installAll(s *session.Session, log io.Writer) error {
	for _, c := range s.Plugins.Clients {
		name := c.Info.Plugin.Name
		if !c.Supports(protocol.MethodInstall) {
			continue
		}
		var projects []protocol.Project
		for _, p := range s.Graph.Projects {
			if p.Plugin == name {
				projects = append(projects, p.Project)
			}
		}
		if len(projects) == 0 {
			continue
		}
		var res protocol.DependencyResult
		if err := c.Call(protocol.MethodInstall, protocol.InstallParams{Projects: projects}, &res, timeout); err != nil {
			return fmt.Errorf("plugin %s: %w", name, err)
		}
		report(log, fmt.Sprintf("%s: installed %d project(s)", name, len(projects)), res)
	}
	return nil
}

// resolve finds each target by project name or workspace-relative
// directory.
func resolve(g *projectgraph.Graph, targets []string) ([]*projectgraph.Project, error) {
	var out []*projectgraph.Project
	for _, t := range targets {
		p := g.Get(t)
		if p == nil {
			if dir, err := projectgraph.CleanRelative(t); err == nil {
				for _, candidate := range g.Projects {
					if candidate.Root == dir {
						p = candidate
					}
				}
			}
		}
		if p == nil {
			return nil, fmt.Errorf("no project %q: name it by its name or its directory (projects: %s).%s", t, strings.Join(g.Names(), ", "), nearmatch.Suggestion(t, g.Names()))
		}
		out = append(out, p)
	}
	return out, nil
}

func outcome(p *projectgraph.Project, supported bool, err error) error {
	if !supported && err == nil {
		return fmt.Errorf("plugin %s cannot change %s's dependencies; edit its manifest instead", p.Plugin, p.Name)
	}
	if err != nil {
		return fmt.Errorf("%s (plugin %s): %w", p.Name, p.Plugin, err)
	}
	return nil
}

func report(log io.Writer, done string, res protocol.DependencyResult) {
	line := "✔ " + done
	if len(res.Files) > 0 {
		line += " (" + strings.Join(res.Files, ", ") + ")"
	}
	fmt.Fprintln(log, line)
	for _, n := range res.Notes {
		fmt.Fprintln(log, "  "+n)
	}
}
