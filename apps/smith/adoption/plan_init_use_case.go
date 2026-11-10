package adoption

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptdecision"
	"github.com/monosmith-app/monosmith/apps/smith/agentfiles"
	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/textdiff"
)

// ProviderPaths maps a CI provider to the file that holds its pipeline.
var ProviderPaths = map[string]string{
	"github": cipipeline.GitHubPath, "azure": cipipeline.AzurePath, "gitlab": cipipeline.GitLabPath,
	"circleci": cipipeline.CircleCIPath, "bitbucket": cipipeline.BitbucketPath,
}

// OwnPipelines maps each provider to the path of a pipeline of the team's own
// at the place smith would write (the file exists and smith did not generate it).
//
// Parameters: root is the repository root.
func OwnPipelines(root string) map[string]string {
	own := map[string]string{}
	for provider, path := range ProviderPaths {
		body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err == nil && !strings.Contains(string(body), GeneratedMarker) {
			own[provider] = path
		}
	}
	return own
}

// InitFacts reads what the repository holds, for the questions of --init.
//
// Parameters: r is the inspection report.
func InitFacts(r Report) adoptdecision.Facts {
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(r.Root, filepath.FromSlash(rel)))
		return err == nil
	}
	return adoptdecision.Facts{
		GitattributesExists: exists(".gitattributes"),
		MCPExists:           exists(".mcp.json"),
		Detected:            r.Detected,
		OwnPipelines:        OwnPipelines(r.Root),
	}
}

// PlanInit is what smith adopt --init will do, before it does it: the
// pipelines of yours it moves aside, the files it creates, and the files of
// yours it edits, each with a diff. It reads and writes nothing.
//
// Parameters: r is the report; c are the settled choices; install is --install.
// Returns: the plan; its Title says "adopt smith".
func PlanInit(r Report, c adoptdecision.Choices, install bool) operationplan.Plan {
	plan := operationplan.Plan{Title: "adopt smith in " + r.Name}
	read := func(rel string) (string, bool) {
		raw, err := os.ReadFile(filepath.Join(r.Root, filepath.FromSlash(rel)))
		return string(raw), err == nil
	}
	providers := adoptdecision.WithoutKept(c, OwnPipelines(r.Root))
	own := OwnPipelines(r.Root)
	for _, p := range providers {
		if path := own[p]; path != "" {
			plan.Add("Pipelines of yours, kept beside smith's", operationplan.Line{Action: operationplan.Change, Path: path, Note: "moved to " + path + ".pre-smith; smith writes its own here"})
		}
	}
	create := func(path, note string) {
		if _, exists := read(path); !exists {
			plan.Add("Files smith creates", operationplan.Line{Action: operationplan.Create, Path: path, Note: note})
		}
	}
	create("smith.toml", "the workspace; "+declaredNote(r))
	create("README.md", "")
	create(".gitignore", "smith's cache is ignored")
	if c.Gitattributes == adoptdecision.GitattributesAdd {
		create(".gitattributes", "`* text=auto eol=lf`: text files become LF")
	}
	for _, p := range providers {
		if path := ProviderPaths[p]; path != "" {
			plan.Add("Files smith creates", operationplan.Line{Action: operationplan.Create, Path: path, Note: p + " pipeline"})
		}
	}
	if len(providers) > 0 {
		plan.Add("Files smith creates", operationplan.Line{Action: operationplan.Create, Path: cipipeline.DevcontainerPath, Note: "dev container with the toolchains CI installs"})
		plan.Add("Files smith creates", operationplan.Line{Action: operationplan.Create, Path: cipipeline.InstallScriptPath, Note: "installs smith in the pipelines"})
	}
	if c.Agent == adoptdecision.AgentAdd {
		files, err := agentfiles.Plan(agentfiles.Options{Root: r.Root, Claude: true, VSCode: true, Agent: true})
		if err == nil {
			for _, f := range files {
				if f.Old == "" {
					plan.Add("Files smith creates", operationplan.Line{Action: operationplan.Create, Path: f.Path, Note: "the MonoSmith agent"})
					continue
				}
				plan.Add("Files of yours smith edits", operationplan.Line{Action: operationplan.Change, Path: f.Path, Note: "merged: what you have stays", Diff: textdiff.UnifiedDiff(f.Path, f.Old, f.New)})
			}
		}
	}
	if c.Agent == adoptdecision.AgentNever {
		plan.Add("Files smith creates", operationplan.Line{Action: operationplan.Create, Path: "smith.toml", Note: "records [agent] enabled = false"})
	}
	edits := func(path, note, before, after string) {
		plan.Add("Files of yours smith edits", operationplan.Line{Action: operationplan.Change, Path: path, Note: note, Diff: textdiff.UnifiedDiff(path, before, after)})
	}
	var nodeRoots, goRoots []string
	for _, p := range r.Projects {
		switch p.Language {
		case "node":
			nodeRoots = append(nodeRoots, p.Root)
		case "go":
			goRoots = append(goRoots, p.Root)
		}
	}
	gitignore, hadIgnore := read(".gitignore")
	ignoreAfter := gitignoreWith(gitignore, len(nodeRoots) > 0)
	if hadIgnore && ignoreAfter != gitignore {
		edits(".gitignore", "smith's cache"+map[bool]string{true: " and node_modules", false: ""}[len(nodeRoots) > 0]+" are ignored", gitignore, ignoreAfter)
	}
	if len(nodeRoots) > 0 {
		before, _ := read("package.json")
		if after, changed, err := withWorkspaces([]byte(before), r.Name, workspacePatterns(nodeRoots)); err == nil && changed {
			edits("package.json", "\"workspaces\" lists the Node projects", before, string(after))
		}
		note := "kept in step with the new workspaces"
		if install {
			note = "npm install runs (it uses the network and can take minutes)"
		}
		if _, hasLock := read("package-lock.json"); hasLock || install {
			plan.Add("Files of yours smith edits", operationplan.Line{Action: operationplan.Change, Path: "package-lock.json", Note: note})
		}
	}
	if len(goRoots) > 0 {
		if _, has := read("go.work"); !has {
			plan.Add("Files smith creates", operationplan.Line{Action: operationplan.Create, Path: "go.work", Note: "go work init and use for the Go projects with a go.mod"})
		}
	}
	return plan
}

func declaredNote(r Report) string {
	if len(r.Declared) > 0 {
		return "it declares the Go projects of the root module"
	}
	return "its projects are found, not listed"
}

// gitignoreWith is the .gitignore after --init: smith's cache folder, and
// node_modules when there are Node projects, each added only when no line
// ignores it yet.
func gitignoreWith(text string, node bool) string {
	has := func(names ...string) bool {
		for _, line := range strings.Split(text, "\n") {
			for _, n := range names {
				if strings.TrimSpace(line) == n {
					return true
				}
			}
		}
		return false
	}
	out := text
	if !has(".smith/", ".smith", "/.smith/", "/.smith") {
		if out != "" && !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		out += "\n# smith\n.smith/\n"
	}
	if node && !has("node_modules", "node_modules/", "/node_modules", "/node_modules/", "**/node_modules", "**/node_modules/") {
		if out != "" && !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		out += "node_modules/\n"
	}
	return out
}
