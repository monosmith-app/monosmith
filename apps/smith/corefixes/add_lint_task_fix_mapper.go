package corefixes

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// AddLintTaskDescription names the fix; every project without a lint task
// shares it, so the host folds them.
const AddLintTaskDescription = "add a lint task that runs ESLint to each project without one"

var eslintConfigs = []string{"eslint.config.mjs", "eslint.config.js", "eslint.config.cjs", "eslint.config.ts", "eslint.config.mts"}

// AddLintTaskFix declares, in smith.toml, a lint task for each of the Node
// projects named that have none: `npx eslint .` in the project, cached on its
// files and the workspace's ESLint config, after its dependencies' builds, as
// the Node plugin does for a package that has a lint script. It is one change
// to smith.toml for all of them, so every project's check carries the same
// fix. It needs an ESLint config at the workspace root and returns nil
// without one, or for a project of another language (smith cannot guess its
// linter). It is not safe: a new task can fail CI on lint errors.
func AddLintTaskFix(root string, projects []string) *protocol.Fix {
	config := ""
	for _, name := range eslintConfigs {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			config = name
			break
		}
	}
	raw, err := os.ReadFile(filepath.Join(root, workspace.FileName))
	if config == "" || err != nil || len(projects) == 0 {
		return nil
	}
	src := string(raw)
	for _, p := range projects {
		table := "projects." + configedit.TableKey(p) + ".tasks.lint"
		src = configedit.SetKey(src, table, "command", `["npx", "eslint", "."]`)
		src = configedit.SetKey(src, table, "cache", "true")
		src = configedit.SetKey(src, table, "inputs", `["**/*", "{workspaceRoot}/`+config+`"]`)
		src = configedit.SetKey(src, table, "depends_on", `["^build"]`)
	}
	return &protocol.Fix{
		Description: AddLintTaskDescription,
		AppliesTo:   projects,
		Files:       []protocol.FileEdit{{Path: workspace.FileName, Content: src}},
	}
}
