package guardedwrite

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/textdiff"
)

// maxDiffLines is how many diff lines a refusal shows.
const maxDiffLines = 20

// WriteFile writes content to path. A file that exists with other content is
// not overwritten unless force is true: the refusal names the file, shows what
// would change, and names the flag. With force the diff is printed, then the
// file is replaced. The same content is left alone.
//
// Parameters: path is the file; content is what to write; force is the flag
// the person passed to allow an overwrite; flag is its name for the message
// (for example "--force"); out gets the notice of a forced overwrite.
// Returns: nil when the file was written or already had the content; a
// refusal when the file exists and differs and force is false; the write error.
func WriteFile(path string, content []byte, force bool, flag string, out io.Writer) error {
	path = filepath.Clean(path)
	old, err := os.ReadFile(path)
	if err == nil {
		if string(old) == string(content) {
			return nil
		}
		diff := textdiff.UnifiedDiff(path, strings.ReplaceAll(string(old), "\r\n", "\n"), strings.ReplaceAll(string(content), "\r\n", "\n"))
		if !force {
			return problem.New(path+" exists already and differs from what smith would write:\n"+head(diff),
				flag+" replaces it", "or write to another file with -o")
		}
		fmt.Fprintf(out, "Replacing %s:\n%s\n", path, head(diff))
	}
	return os.WriteFile(path, content, 0o644)
}

// head is the first lines of a diff, with a count of the rest.
func head(diff string) string {
	lines := strings.Split(strings.TrimRight(diff, "\n"), "\n")
	if len(lines) <= maxDiffLines {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[:maxDiffLines], "\n") + fmt.Sprintf("\n... %d more line(s)", len(lines)-maxDiffLines)
}
