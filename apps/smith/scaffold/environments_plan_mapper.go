package scaffold

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
)

// EnvironmentLines are the plan lines of the environment builds of a web app,
// so the person sees every file and task before smith writes anything: the
// env files, the build options, the git ignore, the scripts and where each
// build goes.
//
// Parameters: dir is the app's workspace-relative folder (apps/web/); s the
// settings, normalized here.
func EnvironmentLines(dir string, s webenv.Settings) []operationplan.Line {
	n := s.Normalized()
	dir = strings.TrimSuffix(dir, "/")
	var lines []operationplan.Line
	for _, env := range n.Names {
		note := "starter VITE_ values; read by `smith run " + webenv.TaskName(n, env) + "`"
		if n.EnvFiles == webenv.FilesExample {
			note += "; kept out of git"
		}
		lines = append(lines, operationplan.Line{Action: operationplan.Create, Path: dir + "/" + webenv.EnvFileName(env), Note: note})
	}
	if n.EnvFiles == webenv.FilesExample {
		lines = append(lines, operationplan.Line{Action: operationplan.Create, Path: dir + "/.env.example", Note: "the one env file git keeps"})
	}
	lines = append(lines,
		operationplan.Line{Action: operationplan.Create, Path: dir + "/vite.build.ts", Note: "build(mode): sourcemaps and minifier of each environment"},
		operationplan.Line{Action: operationplan.Create, Path: dir + "/.gitignore", Note: "keeps .env.*.local private"},
	)
	var tasks []string
	for _, env := range n.Names {
		tasks = append(tasks, fmt.Sprintf("%s -> %s", webenv.TaskName(n, env), strings.TrimPrefix(webenv.OutDir(n, dir, env), "../../")))
	}
	lines = append(lines, operationplan.Line{Action: operationplan.Change, Path: "smith.toml", Note: "records the environments of the app (" + strings.Join(n.Names, ", ") + ") and, if the workspace has no list, the workspace's own"})
	slices.Sort(tasks)
	lines = append(lines, operationplan.Line{Action: operationplan.Create, Path: dir + "/package.json", Note: "one build script per environment, each into its own folder: " + strings.Join(tasks, ", ")})
	return lines
}
