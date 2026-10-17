package discovery

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// NativeTag marks a project built on each OS ("smith": {"native": true}):
// smith ci verify and pack leave it out, smith ci native builds it on every
// runner, and each runner publishes its own platform's build.
const NativeTag = "native"

// workspaceInputs are root files every Node task depends on.
var workspaceInputs = []string{
	"**/*",
	"{workspaceRoot}/package-lock.json",
	"{workspaceRoot}/tsconfig.base.json",
	"{workspaceRoot}/eslint.config.mjs",
}

// scriptTasks describes how each well-known script runs. Scripts not listed
// still become tasks, uncached and without dependencies.
var scriptTasks = map[string]protocol.Task{
	"build":     {DependsOn: []string{"^build"}, Cache: true, Description: "Compile into the workspace's dist/<project dir>"},
	"typecheck": {DependsOn: []string{"^build"}, Cache: true, Description: "Type-check without emitting"},
	"test":      {DependsOn: []string{"^build"}, Cache: true, Description: "Run unit tests"},
	"lint":      {DependsOn: []string{"^build"}, Cache: true, Description: "Lint and check formatting"},
	"start":     {DependsOn: []string{"build"}, Description: "Run the app"},
	"dev":       {Description: "Run the app in watch mode"},
}

// TaskFor builds the task that runs one package.json script through npm, so
// node_modules/.bin is on PATH exactly as when a developer runs it. A
// dependency on another script of the same project (start on build) is kept
// only when the project has that script: plain JavaScript has no build.
func TaskFor(script, root string, scripts map[string]string) protocol.Task {
	t, known := scriptTasks[script]
	if script == "build" {
		// Every build goes to the workspace's dist/<project dir>.
		t.Outputs = []string{protocol.BuildDirFromProject(root)}
	}
	var deps []string
	for _, d := range t.DependsOn {
		if _, has := scripts[d]; has || strings.HasPrefix(d, "^") {
			deps = append(deps, d)
		}
	}
	t.DependsOn = deps
	t.Command = []string{"npm", "run", script, "--silent"}
	if known && t.Cache {
		t.Inputs = workspaceInputs
	}
	if !known {
		t.Description = "npm script " + script
	}
	return t
}

// DepsCheckTask is the name of the task that checks a package declares what
// it imports (smith npm-deps-check).
const DepsCheckTask = "deps-check"

// verifiedBy are the tasks, in order of preference, that a package's
// deps-check runs before, so CI's verify phase (lint, typecheck, test, build)
// runs it: it is the eslint toolchain's lint that a package usually has, and
// the check is not lint's to own.
var verifiedBy = []string{"lint", "build", "test"}

// AddDepsCheck gives a package the deps-check task, cached on its files
// (package.json among them), and makes its first of lint, build or test
// depend on it.
func AddDepsCheck(project *protocol.Project) {
	project.Tasks[DepsCheckTask] = protocol.Task{
		Command:     []string{"{smith}", "npm-deps-check", project.Root},
		Cache:       true,
		Inputs:      []string{"**/*", "{workspaceRoot}/package.json"},
		Description: "Check the package declares every package it imports",
	}
	for _, name := range verifiedBy {
		if t, ok := project.Tasks[name]; ok {
			t.DependsOn = append(t.DependsOn, DepsCheckTask)
			project.Tasks[name] = t
			return
		}
	}
}

// PruneTask writes an app's deploy output into its build directory
// (smith npm-prune): a package.json and lockfile of its production tree, and
// the workspace packages it uses in workspace_modules/, so the directory
// installs on its own with npm ci --omit=dev (a container image does). It
// runs after the app's build and the builds of what it uses, and is not
// cached: it writes into the build's directory.
func PruneTask(root string) protocol.Task {
	return protocol.Task{
		Command:     []string{"{smith}", "npm-prune", root},
		DependsOn:   []string{"build", "^build"},
		Description: "Write the app's deploy output into " + protocol.ProjectBuildDir(root) + ": its production package.json and lockfile, and the workspace packages it uses (workspace_modules/)",
	}
}

// PackageTask packs an app with npm pack into dist/drop, the deliverable
// smith ci pack keeps. An app that builds is packed from its build directory,
// staged with its package.json (smith npm-pack); one that does not, from its
// own directory.
func PackageTask(root string, scripts map[string]string) protocol.Task {
	t := protocol.Task{
		Command:     []string{"npm", "pack", "--pack-destination", "{workspaceRoot}/dist/drop"},
		Description: "Pack the app into dist/drop (smith ci pack)",
	}
	if _, ok := scripts["build"]; ok {
		t.Command = []string{"{smith}", "npm-pack", root}
		t.Description = "Pack the app's build (" + protocol.ProjectBuildDir(root) + ") into dist/drop (smith ci pack)"
		t.DependsOn = []string{"build"}
	}
	return t
}
