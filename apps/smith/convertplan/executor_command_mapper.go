package convertplan

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

// runCommandsOptions are nx:run-commands' own options; any other option is
// forwarded to the command as a --flag, exactly as Nx does.
var runCommandsOptions = map[string]bool{
	"command": true, "commands": true, "cwd": true, "env": true, "envFile": true, "parallel": true,
	"color": true, "args": true, "forwardAllArgs": true, "readyWhen": true, "prefix": true,
	"prefixColor": true, "outputPath": true, "__unparsed__": true, "tty": true, "usePty": true,
	"streamOutput": true, "description": true, "script": true, "maxParallel": true, "bgColor": true,
}

// commandOf reconstructs the command line nx:run-commands runs: the command
// (several only when they run one after another), args, and the options Nx
// forwards. cwd is workspace-relative ("." when Nx runs it at the root).
func commandOf(t nxworkspace.Target) (line, cwd string, env map[string]string, err error) {
	cwd = "."
	var dir string
	if t.Option("cwd", &dir) && dir != "" {
		cwd = path.Clean(strings.ReplaceAll(dir, "\\", "/"))
	}
	t.Option("env", &env)
	line = t.Command
	var single string
	if t.Option("command", &single) {
		line = single
	}
	if raw, ok := t.Options["commands"]; ok {
		var parallel = true
		t.Option("parallel", &parallel)
		var list []json.RawMessage
		if json.Unmarshal(raw, &list) != nil {
			return "", "", nil, fmt.Errorf("commands is not a list")
		}
		if len(list) > 1 && parallel {
			return "", "", nil, fmt.Errorf("it runs %d commands in parallel", len(list))
		}
		var parts []string
		for _, item := range list {
			var s string
			if json.Unmarshal(item, &s) != nil {
				var obj struct {
					Command string `json:"command"`
				}
				_ = json.Unmarshal(item, &obj)
				s = obj.Command
			}
			parts = append(parts, s)
		}
		line = strings.Join(parts, " && ")
	}
	if strings.TrimSpace(line) == "" {
		return "", "", nil, fmt.Errorf("it has no command")
	}
	var args string
	if t.Option("args", &args) && args != "" {
		line += " " + args
	} else {
		var list []string
		if t.Option("args", &list) && len(list) > 0 {
			line += " " + strings.Join(list, " ")
		}
	}
	var extra []string
	for k := range t.Options {
		if !runCommandsOptions[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	for _, k := range extra {
		var v any
		_ = json.Unmarshal(t.Options[k], &v)
		switch val := v.(type) {
		case bool:
			if val {
				line += " --" + k
			} else {
				line += " --" + k + "=false"
			}
		case string, float64:
			line += fmt.Sprintf(" --%s=%v", k, val)
		default:
			return "", "", nil, fmt.Errorf("it forwards the option %q, which is not a plain value", k)
		}
	}
	return line, cwd, env, nil
}

// executorCommand translates a well-known Nx executor into the command it
// stands for, run from the project root. ok is false for an executor with no
// plain equivalent.
func executorCommand(w *nxworkspace.Workspace, p *nxworkspace.Project, t nxworkspace.Target) (line string, needs []string, ok bool) {
	rel := func(option string) string {
		var v string
		if !t.Option(option, &v) || v == "" {
			return ""
		}
		return relativeTo(p.Root, v)
	}
	switch t.Executor {
	case "@nx/esbuild:esbuild":
		main, out := rel("main"), rel("outputPath")
		if main == "" || out == "" {
			return "", nil, false
		}
		platform, format := esbuildTarget(t)
		if !esbuildBundles(t) {
			line, ok := EsbuildTranspile(rel("tsConfig"), main, out, platform, format)
			return line, []string{"esbuild", "typescript"}, ok
		}
		opts := MapEsbuildOptions(t)
		line := fmt.Sprintf("esbuild %s --bundle --packages=external --platform=%s --format=%s --outdir=%s", main, platform, format, out)
		for _, f := range append(opts.Flags, opts.Sourcemap) {
			if f != "" {
				line += " " + f
			}
		}
		return line, []string{"esbuild"}, true
	case "@nx/js:node":
		var build string
		t.Option("buildTarget", &build)
		_, target, _ := strings.Cut(build, ":")
		target, _, _ = strings.Cut(target, ":")
		if bt, found := p.Targets[target]; found {
			var out string
			if bt.Option("outputPath", &out) && out != "" {
				return "node " + relativeTo(p.Root, out) + "/main.js", nil, true
			}
		}
		return "", nil, false
	case "@nx/js:tsc":
		if c := rel("tsConfig"); c != "" {
			return "tsc -p " + c, nil, true
		}
		return "tsc -p tsconfig.lib.json", nil, true
	case "@nx/jest:jest":
		line := "jest"
		if c := rel("jestConfig"); c != "" {
			line += " --config " + c
		}
		var pass bool
		if t.Option("passWithNoTests", &pass) && pass {
			line += " --passWithNoTests"
		}
		return line, nil, true
	case "@nx/eslint:lint", "@nx/linter:eslint":
		return "eslint .", nil, true
	case "@nx/vite:build":
		return "vite build", nil, true
	case "@nx/vite:dev-server":
		return "vite", nil, true
	case "@nx/vite:preview-server":
		return "vite preview", nil, true
	case "@nx/vite:test", "@nx/vitest:test":
		return "vitest run", nil, true
	}
	return "", nil, false
}

// esbuildTarget is an @nx/esbuild build's platform and its first format
// (Nx's defaults: node, cjs).
func esbuildTarget(t nxworkspace.Target) (platform, format string) {
	platform, formats := "node", []string{"cjs"}
	t.Option("platform", &platform)
	t.Option("format", &formats)
	format = "cjs"
	if len(formats) > 0 {
		format = formats[0]
	}
	return platform, format
}

// esbuildBundles reports whether an @nx/esbuild build bundles (Nx's default).
func esbuildBundles(t nxworkspace.Target) bool {
	bundle := true
	t.Option("bundle", &bundle)
	return bundle
}

// executorGap says why a target's executor cannot become a plain command,
// when convert knows more than "no smith equivalent".
func executorGap(t nxworkspace.Target) string {
	if t.Executor == "@nx/esbuild:esbuild" && !esbuildBundles(t) {
		if platform, format := esbuildTarget(t); platform == "node" && format == "cjs" {
			return "with bundle: false for Node CommonJS, Nx also transpiles the workspace libraries the app uses and writes a loader that maps their imports; convert this build by hand (bundling, or tsc)"
		}
	}
	return ""
}

// relativeTo makes a workspace-relative path relative to a project root.
func relativeTo(root, p string) string {
	p = path.Clean(strings.ReplaceAll(p, "\\", "/"))
	p = strings.TrimPrefix(p, "{workspaceRoot}/")
	if p == root {
		return "."
	}
	if strings.HasPrefix(p, root+"/") {
		return strings.TrimPrefix(p, root+"/")
	}
	up := strings.Repeat("../", strings.Count(root, "/")+1)
	return up + p
}
