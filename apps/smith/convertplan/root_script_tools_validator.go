package convertplan

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// toolsWithTheShell are commands a script finds without installing anything:
// they come with Node, npm, the shell or smith itself.
var toolsWithTheShell = map[string]bool{
	"node": true, "npm": true, "npx": true, "smith": true, "corepack": true, "git": true,
	"cd": true, "echo": true, "exit": true, "export": true, "set": true, "test": true, "true": true, "false": true,
	"rm": true, "cp": true, "mv": true, "mkdir": true, "cat": true, "ls": true, "touch": true, "sleep": true,
	"printf": true, "pwd": true, "if": true, "for": true, "while": true, "source": true, ".": true, ":": true,
	"del": true, "copy": true, "move": true, "rmdir": true, "type": true, "call": true, "start": true,
}

// MissingScriptTools returns the programs a script starts that are installed
// neither in root's node_modules/.bin nor on the PATH, in the order the script
// starts them. A workspace without node_modules has none to judge by, so it
// reports none. Only plain command names count: anything a shell would expand
// is left alone.
func MissingScriptTools(root, script string) []string {
	if _, err := os.Stat(filepath.Join(root, "node_modules")); err != nil {
		return nil
	}
	var missing []string
	for _, segment := range strings.Split(strings.NewReplacer("&&", "\n", "||", "\n", ";", "\n", "|", "\n").Replace(script), "\n") {
		words := strings.Fields(segment)
		for len(words) > 0 && strings.Contains(words[0], "=") && !strings.HasPrefix(words[0], "-") {
			words = words[1:]
		}
		if len(words) == 0 {
			continue
		}
		tool := words[0]
		if toolsWithTheShell[tool] || strings.ContainsAny(tool, "$*?{}()<>`\"'\\!~/") || strings.HasPrefix(tool, "-") || installedTool(root, tool) {
			continue
		}
		if !containsString(missing, tool) {
			missing = append(missing, tool)
		}
	}
	return missing
}

func installedTool(root, tool string) bool {
	for _, name := range []string{tool, tool + ".cmd", tool + ".ps1"} {
		if _, err := os.Stat(filepath.Join(root, "node_modules", ".bin", name)); err == nil {
			return true
		}
	}
	_, err := exec.LookPath(tool)
	return err == nil
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
