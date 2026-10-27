package kindchange

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tomltext"
)

const manifestName = "pyproject.toml"

var (
	pathValue      = regexp.MustCompile(`(path[ \t]*=[ \t]*")([^"]+)(")`)
	requirementRef = regexp.MustCompile(`(?m)^(\s*(?:-e|--editable)\s+)(\.{1,2}/\S*)`)
	workspaceTable = regexp.MustCompile(`(?s)(\[tool\.uv\.workspace\].*?members[ \t]*=[ \t]*\[)(.*?)(\])`)
	quotedItem     = regexp.MustCompile(`"([^"]*)"`)
)

// Change answers what making the project params.Project a params.To changes.
// Python has no publish guard in the manifest: what makes a project a package
// is the kind smith reads and the release that uploads it. The rest is the
// move.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	p := params.Project
	edits := fileedits.New(workspaceRoot)
	manifest := path.Join(p.Root, manifestName)
	src, ok := edits.Read(manifest)
	if !ok {
		return protocol.ChangeKindResult{}, fmt.Errorf("%s not found", manifest)
	}
	var result protocol.ChangeKindResult

	src = tomltext.SetString(src, "tool.smith", "kind", string(params.To))
	if params.To == protocol.ProjectPackage {
		if !regexp.MustCompile(`(?m)^[ \t]*license[ \t]*=`).MatchString(src) {
			result.Warnings = append(result.Warnings, manifest+" has no license: PyPI shows it on the project's page")
		}
	}
	moving := params.NewRoot != p.Root
	if moving {
		src = rebasePaths(src, p.Root, params.NewRoot)
		for _, f := range requirementFiles(edits, p.Root) {
			text, _ := edits.Read(f)
			edits.Write(f, requirementRef.ReplaceAllStringFunc(text, func(m string) string {
				parts := requirementRef.FindStringSubmatch(m)
				return parts[1] + pathrefs.Rebase(parts[2], p.Root, params.NewRoot)
			}))
		}
	}
	edits.Write(manifest, src)
	result.Warnings = append(result.Warnings, entryWarnings(workspaceRoot, src, p, params.To)...)

	if moving {
		retargetDependents(edits, params)
		if err := retargetWorkspace(edits, params); err != nil {
			return result, err
		}
		result.Warnings = append(result.Warnings, "run smith install afterwards: the virtual environment installed this project from its old directory")
	}
	result.Files = edits.Edits()
	return result, nil
}

// entryWarnings say what the kind change makes of the project's program.
func entryWarnings(workspaceRoot, manifest string, p protocol.Project, to protocol.ProjectKind) []string {
	_, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "src"))
	hasMain := strings.Contains(manifest, "[project.scripts]") || (err == nil && hasMainModule(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "src")))
	switch {
	case to == protocol.ProjectApp && !hasMain:
		return []string{"an app runs a program, and this project has no __main__.py or [project.scripts]: add one before running it"}
	case to != protocol.ProjectApp && hasMain:
		return []string{fmt.Sprintf("this project has a program (__main__.py or [project.scripts]); as a %s it is installed by other projects and the program is kept but not deployed", to)}
	}
	return nil
}

func hasMainModule(src string) bool {
	found := false
	_ = filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && d.Name() == "__main__.py" {
			found = true
		}
		return nil
	})
	return found
}

func rebasePaths(src, oldRoot, newRoot string) string {
	return pathValue.ReplaceAllStringFunc(src, func(m string) string {
		parts := pathValue.FindStringSubmatch(m)
		return parts[1] + pathrefs.Rebase(parts[2], oldRoot, newRoot) + parts[3]
	})
}

// requirementFiles lists the requirements files directly in a directory.
func requirementFiles(edits *fileedits.Set, root string) []string {
	return edits.List(root, func(name string) bool {
		return strings.HasPrefix(name, "requirements") && strings.HasSuffix(name, ".txt")
	})
}

// retargetDependents points every other project's pyproject.toml and
// requirements files at the new directory.
func retargetDependents(edits *fileedits.Set, params protocol.ChangeKindParams) {
	for _, other := range params.Projects {
		if other.Language != "python" || other.Root == params.Project.Root {
			continue
		}
		files := append([]string{path.Join(other.Root, manifestName)}, requirementFiles(edits, other.Root)...)
		for _, file := range files {
			if text, ok := edits.Read(file); ok {
				updated, _ := pathrefs.Retarget(text, other.Root, params.Project.Root, params.NewRoot)
				edits.Write(file, updated)
			}
		}
	}
}

// retargetWorkspace updates the root files, and lists the new directory in
// the uv workspace's members when none of them covers it.
func retargetWorkspace(edits *fileedits.Set, params protocol.ChangeKindParams) error {
	files := append([]string{manifestName, "uv.lock"}, requirementFiles(edits, "")...)
	for _, file := range files {
		text, ok := edits.Read(file)
		if !ok {
			continue
		}
		text, _ = pathrefs.Retarget(text, "", params.Project.Root, params.NewRoot)
		if file == manifestName {
			text = ensureMember(text, params.NewRoot)
		}
		edits.Write(file, text)
	}
	return nil
}

// ensureMember adds dir to [tool.uv.workspace] members when the table lists
// members and none of them matches dir.
func ensureMember(src, dir string) string {
	loc := workspaceTable.FindStringSubmatchIndex(src)
	if loc == nil {
		return src
	}
	body := src[loc[4]:loc[5]]
	for _, item := range quotedItem.FindAllStringSubmatch(body, -1) {
		if ok, _ := path.Match(item[1], dir); ok {
			return src
		}
	}
	insert := `, "` + path.Dir(dir) + `/*"`
	if strings.TrimSpace(body) == "" {
		insert = `"` + path.Dir(dir) + `/*"`
	} else if strings.Contains(body, "\n") {
		trimmed := strings.TrimRight(body, " \t\r\n")
		trimmed = strings.TrimSuffix(trimmed, ",")
		return src[:loc[4]] + trimmed + `,` + "\n    \"" + path.Dir(dir) + "/*\",\n" + src[loc[5]:]
	}
	return src[:loc[5]] + insert + src[loc[5]:]
}
