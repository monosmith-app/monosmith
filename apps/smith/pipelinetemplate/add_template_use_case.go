package pipelinetemplate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Options choose what Add writes.
type Options struct {
	// Project is the app a deploy template builds.
	Project string
	// Providers default to the workspace's [ci] providers.
	Providers []string
	// Force overwrites a file that exists.
	Force bool
}

// Add writes a template for each provider and returns the files written.
// Every check runs before the first write, so a refusal leaves no file behind.
func Add(s *session.Session, name string, opts Options) ([]string, error) {
	files, err := planTemplate(s, name, opts)
	if err != nil {
		return nil, err
	}
	var written []string
	for _, f := range files {
		if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
			return written, err
		}
		if err := os.WriteFile(f.path, []byte(f.content), 0o644); err != nil {
			return written, err
		}
		written = append(written, f.rel)
	}
	return written, nil
}

// PlanLines are the files Add would write, without writing any.
//
// Parameters: s is the session; name the template; opts what Add takes.
// Returns: one plan line per file (a create, or a change when it is replaced), or the
// error Add would return.
func PlanLines(s *session.Session, name string, opts Options) ([]operationplan.Line, error) {
	files, err := planTemplate(s, name, opts)
	if err != nil {
		return nil, err
	}
	var lines []operationplan.Line
	for _, f := range files {
		action := operationplan.Create
		if _, err := os.Stat(f.path); err == nil {
			action = operationplan.Change
		}
		lines = append(lines, operationplan.Line{Action: action, Path: f.rel, Note: name + " pipeline"})
	}
	return lines, nil
}

// plannedFile is a pipeline file Add will write.
type plannedFile struct{ rel, path, content string }

// planTemplate renders and checks the files of a template; it writes nothing.
func planTemplate(s *session.Session, name string, opts Options) ([]plannedFile, error) {
	t, ok := Find(name)
	if !ok {
		var names []string
		for _, c := range Catalog {
			names = append(names, c.Name)
		}
		return nil, fmt.Errorf("no pipeline template %q (templates: %s)", name, strings.Join(names, ", "))
	}
	app, err := appFor(s, t, opts.Project)
	if err != nil {
		return nil, err
	}
	providers := opts.Providers
	if len(providers) == 0 {
		providers = s.Config.CI.Providers
	}
	contributions, err := cipipeline.Collect(s)
	if err != nil {
		return nil, err
	}
	model := cipipeline.NewModel(contributions, s.Version, s.Config.Release.Branch)
	var files []plannedFile
	for _, provider := range providers {
		if !slices.Contains(t.Providers, provider) {
			if len(opts.Providers) > 0 {
				return nil, fmt.Errorf("%s has no %s form (it has %s)", name, provider, strings.Join(t.Providers, ", "))
			}
			continue
		}
		content, _ := Render(name, provider, model, app)
		rel := Path(name, provider)
		path := filepath.Join(s.Root, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err == nil && !opts.Force {
			return nil, fmt.Errorf("%s exists and is yours; pass --force to replace it", rel)
		} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		files = append(files, plannedFile{rel: rel, path: path, content: content})
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("%s has no form for this workspace's CI (%s); pass --ci %s", name, strings.Join(providers, ", "), t.Providers[0])
	}
	return files, nil
}

func appFor(s *session.Session, t Template, project string) (App, error) {
	if !t.NeedsProject {
		return App{}, nil
	}
	if project == "" {
		return App{}, fmt.Errorf("%s builds one app: pass --project <app>", t.Name)
	}
	p := s.Graph.Get(project)
	if p == nil {
		return App{}, fmt.Errorf("no project %q", project)
	}
	if p.Kind != protocol.ProjectApp {
		return App{}, fmt.Errorf("%s is a %s; %s deploys an app", project, p.Kind, t.Name)
	}
	build, ok := p.Tasks["build"]
	if !ok {
		return App{}, fmt.Errorf("%s has no build task to deploy the output of", project)
	}
	output := "dist"
	if len(build.Outputs) > 0 {
		output = strings.TrimPrefix(filepath.ToSlash(build.Outputs[0]), "./")
	}
	return App{Name: p.Name, Root: p.Root, Output: output}, nil
}
