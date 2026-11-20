package workspacejoin

import (
	"path"
	"regexp"
	"strings"
)

var (
	workspaceTable = regexp.MustCompile(`(?s)(\[tool\.uv\.workspace\].*?members[ \t]*=[ \t]*\[)(.*?)(\])`)
	quotedItem     = regexp.MustCompile(`"([^"]*)"`)
)

// Members are the [tool.uv.workspace] members of a pyproject.toml, and
// whether the file has that list at all.
//
// Parameters: src is the pyproject.toml text.
// Returns: the member globs, and false when there is no uv workspace.
func Members(src string) ([]string, bool) {
	loc := workspaceTable.FindStringSubmatchIndex(src)
	if loc == nil {
		return nil, false
	}
	var out []string
	for _, item := range quotedItem.FindAllStringSubmatch(src[loc[4]:loc[5]], -1) {
		out = append(out, item[1])
	}
	return out, true
}

// Covered says whether a member glob matches the project folder dir.
//
// Parameters: members are the uv workspace members; dir is the project's
// workspace-relative folder.
// Returns: true when the uv workspace includes the project.
func Covered(members []string, dir string) bool {
	for _, m := range members {
		m = strings.TrimSuffix(strings.TrimPrefix(m, "./"), "/")
		if ok, _ := path.Match(m, dir); ok || m == dir {
			return true
		}
	}
	return false
}

// EnsureMember adds the glob of dir's parent folder to [tool.uv.workspace]
// members when the table lists members and none of them matches dir. The rest
// of the file stays as written.
//
// Parameters: src is the root pyproject.toml text; dir is the project's
// workspace-relative folder.
// Returns: the new text, or src when nothing changes.
func EnsureMember(src, dir string) string {
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
