package gitclient

import (
	"strings"
)

// Revision is a commit with its parents and the files it changed.
type Revision struct {
	Commit
	Parents []string
	// Files are the paths the commit changed, relative to the client's folder.
	// A merge commit lists none.
	Files []string
}

// History reads, with one git process, every commit reachable from HEAD and
// not from the commit named by notFrom ("" reads all history), newest first,
// with its parents and changed files. Many questions of the form "which
// commits touched this folder since that tag" are then answered in memory.
//
// Parameters: notFrom is a commit hash, or "" for the whole history.
// Returns: the revisions in git's log order, or the git error.
func (g *Git) History(notFrom string) ([]Revision, error) {
	args := []string{"log", "--name-only", "--relative", "--no-renames", "--format=%x01%H%x1f%P%x1f%an%x1f%s%x1f%b%x02"}
	if notFrom != "" {
		args = append(args, "HEAD", "^"+notFrom)
	} else {
		args = append(args, "HEAD")
	}
	out, err := g.raw(args...)
	if err != nil {
		return nil, err
	}
	var revs []Revision
	for _, chunk := range strings.Split(out, "\x01") {
		head, files, ok := strings.Cut(chunk, "\x02")
		if !ok {
			continue
		}
		parts := strings.SplitN(head, "\x1f", 5)
		if len(parts) < 4 {
			continue
		}
		r := Revision{Commit: Commit{Hash: strings.TrimSpace(parts[0]), Author: parts[2], Subject: parts[3]}, Parents: strings.Fields(parts[1])}
		if len(parts) == 5 {
			r.Body = strings.TrimSpace(parts[4])
		}
		for _, f := range strings.Split(strings.ReplaceAll(files, "\r\n", "\n"), "\n") {
			if f = strings.TrimSpace(f); f != "" {
				r.Files = append(r.Files, f)
			}
		}
		revs = append(revs, r)
	}
	return revs, nil
}

// CommonAncestor is the best common ancestor of all the commits, or "" when
// they share none or git cannot tell (a shallow clone).
func (g *Git) CommonAncestor(commits []string) string {
	if len(commits) == 0 {
		return ""
	}
	if len(commits) == 1 {
		return commits[0]
	}
	out, err := g.Run(append([]string{"merge-base", "--octopus"}, commits...)...)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}
