package adoptflow

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Changes are the files a conversion writes, as the scratch copy shows them
// against the repository: what it creates, edits and removes.
type Changes struct {
	Created []string `json:"created,omitempty"`
	Edited  []string `json:"edited,omitempty"`
	Removed []string `json:"removed,omitempty"`
}

// Total is how many files change.
func (c Changes) Total() int { return len(c.Created) + len(c.Edited) + len(c.Removed) }

// Changes compares the conversion's scratch copy with the repository. It needs
// Prepare to have made the copy.
func (s *Session) Changes() (Changes, error) {
	if s.analysis == nil || s.view == nil {
		return Changes{}, ErrNotConvertible
	}
	var c Changes
	before := map[string]bool{}
	for _, f := range s.analysis.Files {
		before[f] = true
	}
	after := map[string]bool{}
	err := filepath.WalkDir(s.view.Dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(s.view.Dir, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == ".git" || rel == "node_modules" || strings.HasSuffix(rel, "/node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		after[rel] = true
		if !before[rel] {
			c.Created = append(c.Created, rel)
			return nil
		}
		old, err1 := os.ReadFile(filepath.Join(s.Root, filepath.FromSlash(rel)))
		now, err2 := os.ReadFile(p)
		if err1 == nil && err2 == nil && !bytes.Equal(old, now) {
			c.Edited = append(c.Edited, rel)
		}
		return nil
	})
	if err != nil {
		return Changes{}, err
	}
	for f := range before {
		if !after[f] {
			if _, err := os.Stat(filepath.Join(s.Root, filepath.FromSlash(f))); err == nil {
				c.Removed = append(c.Removed, f)
			}
		}
	}
	sort.Strings(c.Created)
	sort.Strings(c.Edited)
	sort.Strings(c.Removed)
	return c, nil
}

// Text lists the changes by what happens to the files, a file that changes in
// several folders on one line.
func (c Changes) Text() string {
	return fmt.Sprintf("● smith will create %d files, edit %d and remove %d\n", len(c.Created), len(c.Edited), len(c.Removed)) + c.Details()
}

// Details lists the files by what happens to them, without the count line
// that Text starts with.
func (c Changes) Details() string {
	var b strings.Builder
	section := func(title string, files []string) {
		if len(files) == 0 {
			return
		}
		fmt.Fprintf(&b, "\n  %s\n", title)
		for _, line := range groupedByName(files) {
			fmt.Fprintf(&b, "    • %s\n", line)
		}
	}
	section("Created", c.Created)
	section("Edited", c.Edited)
	section("Removed", c.Removed)
	return b.String()
}

// groupedByName turns "a/x.json, b/x.json, c/y.json" into
// "x.json in 2 folders" and "c/y.json".
func groupedByName(files []string) []string {
	byName := map[string][]string{}
	var order []string
	for _, f := range files {
		n := path.Base(f)
		if _, ok := byName[n]; !ok {
			order = append(order, n)
		}
		byName[n] = append(byName[n], f)
	}
	var lines []string
	for _, n := range order {
		where := byName[n]
		if len(where) == 1 {
			lines = append(lines, where[0])
			continue
		}
		lines = append(lines, fmt.Sprintf("%s in %d folders", n, len(where)))
	}
	return lines
}
