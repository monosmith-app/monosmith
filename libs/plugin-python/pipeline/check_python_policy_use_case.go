package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
)

// PyPIIndex is the simple index pip uses when a workspace names no registry.
const PyPIIndex = "https://pypi.org/simple"

// CheckPythonPolicy is the Python audit policy over one requirements file:
// every requirement is pinned with a --hash, nothing is fetched from a direct
// URL or a git repository, and the only index is expectedIndex (so a
// dependency cannot come from an index the workspace did not choose). It
// returns one line per violation, in file order.
//
// The workspace's own toolchain file (requirements-dev.txt, written by smith
// with top-level pins only) is exempt from the hash rule: pip's hash-checking
// mode demands a hash for every transitive dependency too, which only a
// resolved lock can give, and the file is smith's, not the project's
// dependencies. Its index and URL rules still apply.
func CheckPythonPolicy(name string, requirements []byte, expectedIndex string) []string {
	if expectedIndex == "" {
		expectedIndex = PyPIIndex
	}
	needHashes := filepath.ToSlash(filepath.Clean(name)) != venv.RequirementsFile
	var violations []string
	for _, line := range logicalLines(string(requirements)) {
		fields := strings.Fields(line)
		switch first := fields[0]; {
		case first == "-i" || first == "--index-url" || strings.HasPrefix(first, "--index-url="):
			if got := optionValue(fields); !sameIndex(got, expectedIndex) {
				violations = append(violations, fmt.Sprintf("%s: index %s is not the expected index %s", name, got, expectedIndex))
			}
		case first == "--extra-index-url" || strings.HasPrefix(first, "--extra-index-url="):
			violations = append(violations, fmt.Sprintf("%s: extra index %s: packages could come from outside the expected index", name, optionValue(fields)))
		case strings.HasPrefix(first, "-"):
			// -r, -c, --find-links...: not a requirement.
		case strings.Contains(line, "@ ") || strings.Contains(line, "git+"):
			violations = append(violations, fmt.Sprintf("%s: %s is fetched from a URL or git, not from the index", name, first))
		case needHashes && !strings.Contains(line, "--hash="):
			violations = append(violations, fmt.Sprintf("%s: %s has no --hash", name, first))
		}
	}
	return violations
}

// CheckPythonPolicyFiles reads each requirements file that exists and checks
// it; a missing file has nothing to check.
func CheckPythonPolicyFiles(paths []string, expectedIndex string) ([]string, error) {
	var violations []string
	for _, p := range paths {
		src, err := os.ReadFile(p)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		violations = append(violations, CheckPythonPolicy(p, src, expectedIndex)...)
	}
	return violations, nil
}

// logicalLines joins backslash continuations and drops comments and blanks.
func logicalLines(src string) []string {
	var lines []string
	var cur strings.Builder
	flush := func() {
		if s := strings.TrimSpace(cur.String()); s != "" {
			lines = append(lines, s)
		}
		cur.Reset()
	}
	for _, raw := range strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n") {
		raw = strings.TrimSpace(raw)
		if strings.HasPrefix(raw, "#") {
			continue
		}
		if i := strings.Index(raw, " #"); i >= 0 {
			raw = strings.TrimSpace(raw[:i])
		}
		if strings.HasSuffix(raw, "\\") {
			cur.WriteString(strings.TrimSuffix(raw, "\\") + " ")
			continue
		}
		cur.WriteString(raw)
		flush()
	}
	flush()
	return lines
}

// optionValue is the value of "--opt value" or "--opt=value".
func optionValue(fields []string) string {
	if _, v, ok := strings.Cut(fields[0], "="); ok {
		return v
	}
	if len(fields) > 1 {
		return fields[1]
	}
	return ""
}

func sameIndex(a, b string) bool {
	return strings.TrimRight(a, "/") == strings.TrimRight(b, "/")
}
