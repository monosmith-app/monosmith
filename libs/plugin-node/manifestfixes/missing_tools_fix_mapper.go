package manifestfixes

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/declaredimports"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// MissingToolsDescription names the fix; the question says which tools.
const MissingToolsDescription = "install the tools the scripts start, or remove those scripts"

// MissingToolsFix asks what to do about scripts that start tools nobody
// installed: install the tools (the package named like the command; a tool
// package.json already declares needs only `npm install`), or remove the
// scripts that start them. byTool maps each missing tool to the scripts of
// the package.json in dir (workspace-relative, "" for the root) that start it.
func MissingToolsFix(workspaceRoot, dir, label string, byTool map[string][]string) *protocol.Fix {
	if len(byTool) == 0 {
		return nil
	}
	manifest := path.Join(dir, "package.json")
	raw, err := readFile(workspaceRoot, manifest)
	if err != nil {
		return nil
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return nil
	}
	tools := make([]string, 0, len(byTool))
	for tool := range byTool {
		tools = append(tools, tool)
	}
	sort.Strings(tools)

	known := declaredimports.WorkspaceSpecs(workspaceRoot)
	var add []string
	declared := false
	for _, tool := range tools {
		if len(known[tool]) > 0 {
			declared = true
			continue
		}
		add = append(add, tool)
	}
	install := protocol.Fix{}
	if declared {
		install.Steps = append(install.Steps, protocol.FixStep{Description: "install what package.json declares", Argv: []string{"npm", "install", "--no-audit", "--no-fund"}})
	}
	if len(add) > 0 {
		argv := []string{"npm", "install", "-D", "--no-audit", "--no-fund"}
		if dir != "" {
			argv = append(argv, "--workspace", dir)
		}
		install.Steps = append(install.Steps, protocol.FixStep{Description: "add " + strings.Join(add, ", ") + " to devDependencies", Argv: append(argv, add...)})
	}

	var removed []string
	scripts := doc.Object("scripts", false)
	for _, tool := range tools {
		for _, script := range byTool[tool] {
			if scripts != nil && scripts.Has(script) {
				command, _ := scripts.String(script)
				removed = append(removed, fmt.Sprintf("%s (%s)", script, command))
				scripts.Delete(script)
			}
		}
	}
	remove := protocol.Fix{Files: []protocol.FileEdit{{Path: manifest, Content: render(doc, raw)}}}

	question := fmt.Sprintf("%s starts %s, which is not installed.", label, strings.Join(tools, ", "))
	var appliesTo []string
	if dir != "" {
		appliesTo = []string{label}
	}
	return &protocol.Fix{
		Description: MissingToolsDescription,
		AppliesTo:   appliesTo,
		Choices: []protocol.FixChoice{{
			Question: question,
			Options: []protocol.FixChoiceOption{
				{Label: "install " + strings.Join(tools, ", ") + " (the package named like the command)", Fix: install},
				{Label: "remove the script(s): " + strings.Join(removed, "; "), Fix: remove},
			},
		}},
	}
}

func readFile(workspaceRoot, rel string) ([]byte, error) {
	return os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
}
