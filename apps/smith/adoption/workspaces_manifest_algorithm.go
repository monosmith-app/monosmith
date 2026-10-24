package adoption

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"sort"
)

// workspacePatterns are the root package.json "workspaces" entries that cover
// the given project roots: a project at the top is named, others are covered
// by their parent folder (packages/*).
func workspacePatterns(roots []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, root := range roots {
		pattern := root
		if dir := path.Dir(root); dir != "." {
			pattern = dir + "/*"
		}
		if !seen[pattern] {
			seen[pattern] = true
			out = append(out, pattern)
		}
	}
	sort.Strings(out)
	return out
}

// withWorkspaces returns a root package.json that has the workspaces entry,
// and says whether it changed. A manifest that already has a "workspaces" key
// is left alone: merging into the user's own list is theirs to do. Everything
// else in the manifest stays exactly as written. An empty raw starts a new
// private manifest named after the workspace.
func withWorkspaces(raw []byte, workspaceName string, patterns []string) ([]byte, bool, error) {
	list, _ := json.Marshal(patterns)
	list = bytes.ReplaceAll(list, []byte(`","`), []byte(`", "`))
	if len(bytes.TrimSpace(raw)) == 0 {
		return []byte(fmt.Sprintf("{\n  \"name\": %q,\n  \"private\": true,\n  \"workspaces\": %s\n}\n", workspaceName+"-workspace", list)), true, nil
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil {
		return raw, false, fmt.Errorf("package.json: %w", err)
	}
	if _, has := probe["workspaces"]; has {
		return raw, false, nil
	}
	body := bytes.TrimRight(raw, " \t\r\n")
	end := bytes.LastIndexByte(body, '}')
	if end < 0 {
		return raw, false, fmt.Errorf("package.json has no closing brace")
	}
	head := bytes.TrimRight(body[:end], " \t\r\n")
	sep := ",\n  "
	if bytes.HasSuffix(head, []byte("{")) {
		sep = "\n  "
	}
	out := string(head) + sep + `"workspaces": ` + string(list) + "\n}\n"
	if err := json.Unmarshal([]byte(out), &probe); err != nil {
		return raw, false, fmt.Errorf("package.json: could not add workspaces: %w", err)
	}
	return []byte(out), true, nil
}
