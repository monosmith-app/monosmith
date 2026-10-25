package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// pluginName is the plugin's name, and the key of its settings.
const pluginName = "widget-bundle"

// defaultTarget is where a release deploys when no target is configured.
const defaultTarget = "dist/deploy"

// bundle is a deploy plugin. Contribute puts the bundle task on apps, which
// is also what makes smith treat the plugin as the app's deployer; Publish is
// called on release; CI says what the pipeline needs; Doctor checks the
// settings.
type bundle struct{ root string }

// Describe answers "initialize".
func (b *bundle) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	b.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: pluginName, Version: "0.1.0"},
		Kind:      protocol.KindConvention,
		Roles:     []protocol.Role{protocol.RoleDeploy},
		Languages: []string{"widget"},
	}, nil
}

// targetOf reads the plugin's settings for a project, which smith takes from
// smith.toml and hands over in Project.Settings["widget-bundle"]:
//
//	[projects.shop.settings.widget-bundle]
//	target = "dist/deploy"
//
// ([conventions.<name>] tables carry only the fields smith knows, which are
// the docker convention's; an external plugin is configured per project.)
func targetOf(p protocol.Project) string {
	if t, _ := p.Settings[pluginName]["target"].(string); t != "" {
		return t
	}
	return defaultTarget
}

// bundleDir is where the bundle task writes, inside dist/ like every build.
func bundleDir(root string) string { return protocol.ProjectBuildDir(root) + "-bundle" }

// Contribute answers "contribute": a bundle task for every app.
func (b *bundle) Contribute(p protocol.ContributeParams) (protocol.ContributeResult, error) {
	exe, err := os.Executable()
	if err != nil {
		return protocol.ContributeResult{}, err
	}
	out := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, project := range p.Projects {
		if project.Kind != protocol.ProjectApp {
			continue // a lib is not deployed
		}
		out.Projects[project.Name] = protocol.Contribution{Tasks: map[string]protocol.Task{
			"bundle": {
				Command:     []string{exe, "bundle", protocol.BuildDirFromProject(project.Root), protocol.BuildDirFromProject(project.Root) + "-bundle"},
				Inputs:      []string{"widget.json"},
				Outputs:     []string{protocol.BuildDirFromProject(project.Root) + "-bundle"},
				DependsOn:   []string{"build"},
				Cache:       true,
				Description: "bundle the app's build for deployment to " + targetOf(project),
			},
		}}
	}
	return out, nil
}

// Publish answers "publish": deploy one app at one version. It is called for
// the apps this plugin contributed a task to, after their language plugin
// published them. It must be safe to run again: an already deployed version
// is Skipped with a reason, which is a success.
func (b *bundle) Publish(p protocol.PublishParams) (protocol.PublishResult, error) {
	target := targetOf(p.Project)
	file := fmt.Sprintf("%s-%s.txt", p.Project.Name, p.Version)
	dest := path.Join(target, file)
	if _, err := os.Stat(filepath.Join(b.root, filepath.FromSlash(dest))); err == nil {
		return protocol.PublishResult{Skipped: dest + " is already deployed", Registry: target}, nil
	}
	if p.DryRun {
		return protocol.PublishResult{Skipped: "dry run: would deploy to " + dest, Registry: target}, nil
	}
	// Release builds the app (its build task) but not the tasks this plugin
	// added, so the deploy makes the bundle itself from the build output.
	src := filepath.Join(b.root, filepath.FromSlash(protocol.ProjectBuildDir(p.Project.Root)), "widget.txt")
	built, err := os.ReadFile(src)
	if err != nil {
		// A failure the user must act on: say what to do.
		return protocol.PublishResult{}, fmt.Errorf("no build output at %s: run smith run build -p %s first", filepath.ToSlash(src), p.Project.Name)
	}
	content := bundleOf(built)
	full := filepath.Join(b.root, filepath.FromSlash(dest))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return protocol.PublishResult{}, err
	}
	if err := os.WriteFile(full, content, 0o644); err != nil {
		return protocol.PublishResult{}, err
	}
	return protocol.PublishResult{Published: true, Registry: target}, nil
}

// CI answers "ci": the credential a real deploy would need, and the step
// that runs after every release.
func (b *bundle) CI(p protocol.CIParams) (protocol.CIResult, error) {
	var res protocol.CIResult
	for _, project := range p.Projects {
		if project.Kind == protocol.ProjectApp {
			res.Credentials = []protocol.Credential{{
				Env:         "WIDGET_DEPLOY_TOKEN",
				Description: "token of the widget deploy target (this example's deploy ignores it)",
				Required:    false,
			}}
			break
		}
	}
	return res, nil
}

// Doctor answers "doctor": a target must stay inside the workspace.
func (b *bundle) Doctor(p protocol.DoctorParams) (protocol.DoctorResult, error) {
	result := protocol.DoctorResult{Findings: []protocol.Finding{}}
	for _, project := range p.Projects {
		if project.Kind != protocol.ProjectApp {
			continue
		}
		target := targetOf(project)
		ok := !outside(target)
		f := protocol.Finding{Check: project.Name + " deploys inside the workspace", OK: ok}
		if !ok {
			f.Detail = "the deploy target " + target + " is outside the workspace"
			f.Remedy = "set target in [projects." + project.Name + ".settings." + pluginName + "] in smith.toml to a workspace-relative folder"
		}
		result.Findings = append(result.Findings, f)
	}
	return result, nil
}

// outside reports whether a workspace-relative path leaves the workspace.
func outside(p string) bool {
	if filepath.IsAbs(p) {
		return true
	}
	return slices.Contains(strings.Split(filepath.ToSlash(p), "/"), "..")
}

// bundleOf is what the bundle task and the deploy both produce from a build.
func bundleOf(built []byte) []byte { return append([]byte("bundle of "), built...) }
