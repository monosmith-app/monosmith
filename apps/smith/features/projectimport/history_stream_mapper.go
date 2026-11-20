package projectimport

import (
	"fmt"
	"strings"
)

// LogFormat is the git log format HistoryStream reads: the fields of one
// commit separated by NUL, each commit ended by \x01. Dates must be raw
// (git log --date=raw).
const LogFormat = "%H%x00%P%x00%T%x00%an%x00%ae%x00%ad%x00%cn%x00%ce%x00%cd%x00%B%x01"

// emptyTree is the hash of git's empty tree.
const emptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// HistoryStream turns the output of git log --reverse --topo-order
// --date=raw --format=LogFormat into a git fast-import stream that writes
// the same commits on branch, with every commit's files under folder. The
// authors, the dates and the messages stay; each commit's tree is the old
// tree placed at folder, so git log -- <folder> and git blame follow the
// history in the workspace.
//
// Parameters: log is git log's output; branch is the ref to write
// (refs/heads/...); folder is the project's folder with forward slashes.
// Returns: the stream, and the number of commits in it.
func HistoryStream(log, branch, folder string) (string, int) {
	marks := map[string]int{}
	var b strings.Builder
	n := 0
	for _, record := range strings.Split(log, "\x01") {
		record = strings.TrimLeft(record, "\r\n")
		f := strings.SplitN(record, "\x00", 10)
		if len(f) < 10 {
			continue
		}
		n++
		marks[f[0]] = n
		fmt.Fprintf(&b, "commit %s\nmark :%d\n", branch, n)
		fmt.Fprintf(&b, "author %s <%s> %s\n", f[3], f[4], f[5])
		fmt.Fprintf(&b, "committer %s <%s> %s\n", f[6], f[7], f[8])
		fmt.Fprintf(&b, "data %d\n%s\n", len(f[9]), f[9])
		first := true
		for _, parent := range strings.Fields(f[1]) {
			mark, ok := marks[parent]
			if !ok {
				continue
			}
			if first {
				fmt.Fprintf(&b, "from :%d\n", mark)
				first = false
			} else {
				fmt.Fprintf(&b, "merge :%d\n", mark)
			}
		}
		b.WriteString("deleteall\n")
		if f[2] != emptyTree {
			fmt.Fprintf(&b, "M 040000 %s %s\n", f[2], folder)
		}
		b.WriteString("\n")
	}
	return b.String(), n
}
