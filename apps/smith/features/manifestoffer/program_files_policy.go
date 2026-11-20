package manifestoffer

import (
	"path"
	"strings"
)

// programFiles are the base names of the files that make a folder of
// sources look like a program.
var programFiles = map[string]bool{
	"main.ts": true, "main.js": true, "server.ts": true, "server.js": true, "cli.ts": true, "cli.js": true,
	"main.py": true, "__main__.py": true, "app.py": true, "program.cs": true,
}

// LooksLikeProgram says whether a folder of sources looks like a program: a
// main, server, cli or Program file at its top. The answer is only the
// suggested kind: the person chooses it.
//
// Parameters: files are the source files, relative, with forward slashes.
// Returns: true for a program.
func LooksLikeProgram(files []string) bool {
	for _, f := range files {
		if !strings.Contains(f, "/") && programFiles[strings.ToLower(path.Base(f))] {
			return true
		}
	}
	return false
}
