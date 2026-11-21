package outputignore

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
)

// probeFile is the name of the file Missing asks git about inside a folder
// pattern: git judges the path, so the file need not exist.
const probeFile = "smith-probe"

// Missing returns the .gitignore lines that make git ignore each pattern in
// folder, for the patterns git does not ignore there yet. Each line is
// anchored to the folder (/packages/astar/bin/), so a pattern such as bin/
// never hides a bin/ folder of another project. Outside a git repository
// every pattern counts as not ignored.
//
// Parameters: root is the workspace root; folder is the project folder from
// the root, with forward slashes; patterns are folder-relative gitignore
// patterns (bin/, *.egg-info/).
// Returns: the lines to add, in the order of patterns; empty when git ignores
// every pattern.
func Missing(root, folder string, patterns []string) []string {
	if len(patterns) == 0 {
		return nil
	}
	folder = strings.Trim(filepath.ToSlash(folder), "/")
	probes := make([]string, len(patterns))
	for i, p := range patterns {
		probes[i] = probe(folder, p)
	}
	out, _ := gitclient.New(root).Run(append([]string{"check-ignore", "--"}, probes...)...)
	ignored := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		ignored[strings.TrimSpace(filepath.ToSlash(line))] = true
	}
	var lines []string
	for i, p := range patterns {
		if !ignored[probes[i]] {
			lines = append(lines, "/"+folder+"/"+strings.TrimPrefix(p, "/"))
		}
	}
	return lines
}

// probe is a path git ignores exactly when it ignores pattern in folder: the
// pattern with * as x, and a file inside it when the pattern is a folder.
func probe(folder, pattern string) string {
	name := strings.ReplaceAll(strings.TrimPrefix(pattern, "/"), "*", "x")
	if strings.HasSuffix(name, "/") {
		name += probeFile
	}
	return folder + "/" + name
}

// Append adds lines to the workspace .gitignore under one comment that names
// the folder, and creates the file when there is none. The other lines stay
// as they are.
//
// Parameters: root is the workspace root; folder is the project folder;
// lines are the lines Missing returned.
// Returns: an error when .gitignore cannot be read or written.
func Append(root, folder string, lines []string) error {
	if len(lines) == 0 {
		return nil
	}
	path := filepath.Join(root, ".gitignore")
	current, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(With(string(current), folder, lines)), 0o644)
}

// With is the text of a .gitignore after Append: the current text, a blank
// line, a comment that names the folder, then the lines.
//
// Parameters: current is the text of .gitignore ("" when there is none);
// folder is the project folder; lines are the lines to add.
// Returns: the new text, ending in a newline.
func With(current, folder string, lines []string) string {
	var b strings.Builder
	b.WriteString(current)
	if current != "" && !strings.HasSuffix(current, "\n") {
		b.WriteString("\n")
	}
	if current != "" {
		b.WriteString("\n")
	}
	b.WriteString("# build output of " + strings.Trim(folder, "/") + "\n")
	for _, l := range lines {
		b.WriteString(l + "\n")
	}
	return b.String()
}
