package discovery

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// eslintConfigFiles are the files that configure ESLint in a project folder:
// the flat config of ESLint 9 and the legacy .eslintrc of ESLint 8 and older.
var eslintConfigFiles = []string{
	"eslint.config.js", "eslint.config.mjs", "eslint.config.cjs", "eslint.config.ts", "eslint.config.mts", "eslint.config.cts",
	".eslintrc", ".eslintrc.js", ".eslintrc.cjs", ".eslintrc.json", ".eslintrc.yml", ".eslintrc.yaml",
}

// InferredLintTask gives a project a lint task when it has ESLint configured
// but no lint script: the project's own ESLint, run on its folder. Without it,
// the project's lint setup is there but nothing runs it.
//
// Parameters: dir is the project's folder on disk; pkg is its package.json.
// Returns: the task, and true when the project has no lint script, declares
// eslint in its dependencies or devDependencies, and has an ESLint config
// file in dir or an eslintConfig key in package.json.
func InferredLintTask(dir string, pkg PackageManifest) (protocol.Task, bool) {
	if _, own := pkg.Scripts["lint"]; own {
		return protocol.Task{}, false
	}
	_, dep := pkg.Dependencies["eslint"]
	_, dev := pkg.DevDependencies["eslint"]
	if !dep && !dev {
		return protocol.Task{}, false
	}
	configured := len(pkg.EslintConfig) > 0 && string(pkg.EslintConfig) != "null"
	for _, name := range eslintConfigFiles {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			configured = true
			break
		}
	}
	if !configured {
		return protocol.Task{}, false
	}
	t := scriptTasks["lint"]
	t.DependsOn = []string{"^build"}
	t.Command = []string{"npx", "eslint", "."}
	t.Description = "Lint with the project's ESLint config (no lint script)"
	t.Inputs = workspaceInputs
	t.Fix = eslintFix
	return t, true
}
