package toolinstall

import (
	"os"
	"path/filepath"
)

// projectHasFile reports whether a project needs a tool that its declaration
// ties to files: the project folder holds one of the files. A tool with no
// files is needed by every project. A project with no known folder cannot be
// checked, so it does not need a tool tied to files.
//
// Parameters: root is the absolute folder of the project; files are the file
// names from the tool's declaration.
func projectHasFile(root string, files []string) bool {
	if len(files) == 0 {
		return true
	}
	if root == "" {
		return false
	}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(root, f)); err == nil {
			return true
		}
	}
	return false
}
