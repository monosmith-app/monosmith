package kindchange

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Options configure smith kind.
type Options struct {
	Name string
	To   protocol.ProjectKind
	// DryRun says what would change, changing nothing.
	DryRun bool
	// Force goes ahead against the reasons a person may overrule.
	Force bool
}

// Plan is everything a kind change does, worked out and not yet done.
type Plan struct {
	Project string               `json:"project"`
	Plugin  string               `json:"plugin"`
	From    protocol.ProjectKind `json:"from"`
	To      protocol.ProjectKind `json:"to"`
	Root    string               `json:"root"`
	// NewRoot is the folder the project ends up in; equal to Root when it stays.
	NewRoot string `json:"newRoot"`
	// Edits are the file changes, at the paths the files have before the move.
	Edits []protocol.FileEdit `json:"edits,omitempty"`
	// Warnings are what the person should know and decide on.
	Warnings []string `json:"warnings,omitempty"`
	// Mentions are files that still name the old folder, which smith does not
	// know how to rewrite.
	Mentions []string `json:"mentions,omitempty"`
	// Overruled are the reasons --force went past.
	Overruled []string `json:"overruled,omitempty"`
}

// Moves says whether the project's folder changes.
func (p Plan) Moves() bool { return p.Root != p.NewRoot }

// PlanChange works out what making a project another kind involves, and
// refuses what would break.
func PlanChange(s *session.Session, opts Options) (Plan, error) {
	p := s.Graph.Get(opts.Name)
	if p == nil {
		return Plan{}, problem.New(fmt.Sprintf("no project named %q", opts.Name), "smith projects lists them")
	}
	newRoot, _ := projectkind.Relocate(p.Root, opts.To)
	plan := Plan{Project: p.Name, Plugin: p.Plugin, From: p.Kind, To: opts.To, Root: p.Root, NewRoot: newRoot}

	subject := Subject{
		Name: p.Name, Language: p.Language, From: p.Kind, To: opts.To, Root: p.Root, NewRoot: newRoot,
		Dependents:         related(s.Graph, s.Graph.Dependents(p.Name)),
		Dependencies:       related(s.Graph, p.Deps),
		HasReleaseSettings: p.Release != nil,
	}
	if plan.Moves() {
		_, err := os.Stat(filepath.Join(s.Root, filepath.FromSlash(newRoot)))
		subject.TargetExists = err == nil
	}
	if _, tag, found := release.LatestTag(s.Git, p); found {
		subject.LatestTag = tag
	}
	for _, f := range s.Files {
		if f == path.Join(p.Root, "Dockerfile") {
			subject.HasDockerfile = true
		}
	}
	refusals, warnings := Review(subject)
	plan.Warnings = warnings
	if err := refuse(p.Name, opts, refusals, &plan); err != nil {
		return plan, err
	}

	client := s.Plugins.Get(p.Plugin)
	if client == nil || !client.Supports(protocol.MethodChangeKind) {
		return plan, problem.New(
			fmt.Sprintf("the %s plugin cannot change a project's kind yet", p.Plugin),
			fmt.Sprintf("edit %s's manifest and move its folder by hand, then run smith doctor", p.Name),
		)
	}
	var answer protocol.ChangeKindResult
	params := protocol.ChangeKindParams{Project: p.Project, To: opts.To, NewRoot: newRoot, Projects: projectsOf(s.Graph)}
	if err := client.Call(protocol.MethodChangeKind, params, &answer, 5*time.Minute); err != nil {
		return plan, fmt.Errorf("plugin %s: %w", p.Plugin, err)
	}
	if answer.Refused != "" {
		return plan, problem.New(fmt.Sprintf("%s cannot become %s: %s", p.Name, article(opts.To), answer.Refused))
	}
	plan.Edits = answer.Files
	plan.Warnings = append(plan.Warnings, answer.Warnings...)

	if plan.Moves() {
		config, err := configEdit(s.Root, plan)
		if err != nil {
			return plan, err
		}
		if config != nil {
			plan.Edits = append(plan.Edits, *config)
		}
		plan.Mentions = remainingMentions(s.Root, s.Files, plan)
	}
	return plan, nil
}

// refuse turns the review's refusals into the error a person reads, unless
// --force overrules every one of them.
func refuse(name string, opts Options, refusals []Refusal, plan *Plan) error {
	var blocking, forceable []string
	for _, r := range refusals {
		if r.Forceable {
			forceable = append(forceable, r.Reason)
		} else {
			blocking = append(blocking, r.Reason)
		}
	}
	if len(blocking) == 0 && (len(forceable) == 0 || opts.Force) {
		plan.Overruled = forceable
		return nil
	}
	reasons := append(blocking, forceable...)
	text := fmt.Sprintf("%s cannot become %s: %s", name, article(opts.To), strings.Join(reasons, "; "))
	if len(blocking) == 0 {
		return problem.New(text, fmt.Sprintf("smith kind %s %s --force goes ahead anyway; add --dry-run to see what it would change", name, opts.To))
	}
	return problem.New(text)
}

func related(g *projectgraph.Graph, names []string) []Related {
	out := make([]Related, 0, len(names))
	for _, n := range names {
		if p := g.Get(n); p != nil {
			out = append(out, Related{Name: p.Name, Kind: p.Kind, Language: p.Language})
		}
	}
	return out
}

func projectsOf(g *projectgraph.Graph) []protocol.Project {
	out := make([]protocol.Project, len(g.Projects))
	for i, p := range g.Projects {
		// The plugin needs where projects are, not their tasks.
		out[i] = protocol.Project{Name: p.Name, Root: p.Root, Language: p.Language, Kind: p.Kind, Exports: p.Exports}
	}
	return out
}

// configEdit rewrites the paths smith.toml names inside the project's folder
// (its schema files, a Docker build context).
func configEdit(root string, plan Plan) (*protocol.FileEdit, error) {
	raw, err := os.ReadFile(filepath.Join(root, workspace.FileName))
	if err != nil {
		return nil, err
	}
	text, n := pathrefs.Retarget(string(raw), "", plan.Root, plan.NewRoot)
	if n == 0 {
		return nil, nil
	}
	if _, err := workspace.Parse([]byte(text)); err != nil {
		return nil, fmt.Errorf("refusing to write an invalid %s: %w", workspace.FileName, err)
	}
	return &protocol.FileEdit{Path: workspace.FileName, Content: text}, nil
}
