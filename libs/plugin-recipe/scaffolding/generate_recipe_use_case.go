package scaffolding

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

// A recipe's name is its project's name and starts its release tags.
var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

type templateData struct{ Dir, Name string }

// templates are a new recipe: a language choice that resolves, to edit.
var templates = map[string]string{
	"[[.Dir]]/[[.Name]].recipe.toml": `# The recipe [[.Name]]. Edit the choices and the plugins, or write it again
# with ` + "`smith recipe new`" + `. ` + "`smith recipe check [[.Name]].recipe.toml --all`" + ` checks it for
# every answer; releasing it (conventional commits, as for any package) tags
# [[.Name]]@x.y.z and attaches this file to the release.
name = "[[.Name]]"
version = "0.0.0"

[choices.language]
ask = "Language"
options = ["node", "python"]
default = "node"

[when.language.node.use]
language = "node"

[when.language.python.use]
language = "python"
`,
}

// Generate writes a recipe project, whatever kind is asked: a recipe is
// released, so it is a package.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if !namePattern.MatchString(params.Name) {
		return protocol.GenerateResult{}, fmt.Errorf("%q is not a recipe name: use lowercase words joined by dashes (team, team-node), since it starts the release tags", params.Name)
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	written, err := templating.Write(workspaceRoot, templates, templateData{Dir: params.Directory, Name: params.Name})
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	note := "Check it with smith recipe check " + params.Directory + "/" + params.Name + ".recipe.toml --all, then use it with [workspace] recipe in smith.toml."
	return protocol.GenerateResult{Files: written, Notes: []string{note}}, nil
}
