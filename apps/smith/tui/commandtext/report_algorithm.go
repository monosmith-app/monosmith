package commandtext

import (
	"os"
	"path/filepath"
	"strings"
)

// ReportPath is where the report of a workspace is written: a file in the
// temporary folder, named after the workspace (header is its name).
func ReportPath(header string) string {
	name := strings.Map(func(r rune) rune {
		if r == '-' || r == '_' || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return r
		}
		return '-'
	}, header)
	if strings.Trim(name, "-") == "" {
		name = "workspace"
	}
	return filepath.Join(os.TempDir(), "smith-report-"+name+".html")
}

// ReportTarget is the file a `smith report -o <file>` command line writes, and
// whether argv is such a line.
func ReportTarget(argv []string) (string, bool) {
	if len(argv) == 0 || argv[0] != "report" {
		return "", false
	}
	for i, a := range argv {
		if (a == "-o" || a == "--output") && i+1 < len(argv) {
			return argv[i+1], true
		}
	}
	return "", false
}
