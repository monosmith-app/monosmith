package removeproject

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

// shortNameLimit is the longest project name that is matched by its root path only:
// a name like "web" or "api" is too common as a whole word.
const shortNameLimit = 4

// lockfilePathspecs are the tracked files never scanned: generated lists of packages.
var lockfilePathspecs = []string{
	":(exclude,glob)**/*.lock", ":(exclude,glob)**/*.lockb", ":(exclude,glob)**/package-lock.json",
	":(exclude,glob)**/npm-shrinkwrap.json", ":(exclude,glob)**/pnpm-lock.yaml", ":(exclude,glob)**/go.sum",
}

// FindMentions lists the tracked files of the repository at dir that still name a
// removed project, by whole-word name and by former root path. Names of 4
// characters or fewer match the root path only (rootOnly is then true). It skips
// binary files, lockfiles, the removed directory and the files in skip.
// It returns nothing when git is missing or dir is not a repository: this scan
// never fails the command. Each file appears once, with its first matching line.
//
// Parameters: dir is the workspace root; name and projectRoot identify the project; skip is the set of files smith changed.
func FindMentions(dir, name, projectRoot string, skip []string) (mentions []Mention, rootOnly bool) {
	rootOnly = len([]rune(name)) <= shortNameLimit
	first := map[string]int{}
	scan := func(term string, word bool) {
		args := []string{"grep", "-n", "-I", "-F", "--no-color"}
		if word {
			args = append(args, "-w")
		}
		args = append(args, "-e", term, "--", ".", ":(exclude)"+projectRoot)
		args = append(args, lockfilePathspecs...)
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, _ := cmd.Output() // exit 1 is "no match"; any other failure means no result
		for _, line := range bytes.Split(out, []byte("\n")) {
			parts := strings.SplitN(string(line), ":", 3)
			if len(parts) < 3 {
				continue
			}
			if !word && !hasPathBoundary(parts[2], term) {
				continue
			}
			n, err := strconv.Atoi(parts[1])
			if err != nil {
				continue
			}
			if old, ok := first[parts[0]]; !ok || n < old {
				first[parts[0]] = n
			}
		}
	}
	scan(projectRoot, false)
	if !rootOnly {
		scan(name, true)
	}
	skipped := map[string]bool{}
	for _, f := range skip {
		skipped[f] = true
	}
	for f, n := range first {
		if !skipped[f] {
			mentions = append(mentions, Mention{File: f, Line: n, Kind: mentionKind(f)})
		}
	}
	return GroupMentions(mentions), rootOnly
}
