// Package copymeasure counts the files that several projects keep as exact
// copies of each other (a tsconfig, a lint config, a jest config) and the
// lines they cost, so a team can see how much of a workspace could live once
// at the workspace root (ADR 0059). It reports; it deletes nothing.
package copymeasure

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"path"
	"sort"
	"strings"
)

// MinCopies is how many projects must hold the same file before it is a copy
// worth naming.
const MinCopies = 3

// maxBytes is the largest file looked at: a config file is small.
const maxBytes = 64 * 1024

// Project is one project: its name, its folder and its files.
type Project struct {
	Name, Root string
	// Files are workspace-relative paths of the files under Root.
	Files []string
}

// Copy is one file name that several projects hold with the same content.
type Copy struct {
	// File is the path inside the project ("tsconfig.json").
	File string `json:"file"`
	// Projects hold the same content, sorted.
	Projects []string `json:"projects"`
	// Lines is the number of lines of one copy.
	Lines int `json:"lines"`
	// Different are the projects that hold this file with other content, sorted.
	Different []string `json:"different,omitempty"`
}

// Report is the result of a measure.
type Report struct {
	Copies []Copy `json:"copies"`
	// LinesNow is the lines of all the copies; LinesOnce is the lines if each
	// copy lived once. ConfigLines is the lines of every file at a project's
	// root, of any content, to put the first in proportion.
	LinesNow    int `json:"linesNow"`
	LinesOnce   int `json:"linesOnce"`
	ConfigLines int `json:"configLines"`
}

// Measure finds the files at the root of projects that hold the same content
// in MinCopies or more projects. Line ends and trailing white space do not
// count as a difference. Only files directly in a project's folder count: they
// are the per-project config.
//
// Parameters: projects lists the projects and their files; read returns a
// file's bytes (workspace-relative path).
// Returns the report, copies largest saving first.
func Measure(projects []Project, read func(string) ([]byte, error)) Report {
	type holder struct{ project, hash string }
	byName := map[string][]holder{}
	lines := map[string]int{} // hash -> lines
	var r Report
	for _, p := range projects {
		for _, f := range p.Files {
			if path.Dir(f) != p.Root && !(p.Root == "." && !strings.Contains(f, "/")) {
				continue
			}
			data, err := read(f)
			if err != nil || len(data) > maxBytes || bytes.IndexByte(data, 0) >= 0 {
				continue
			}
			norm, n := normalise(data)
			r.ConfigLines += n
			sum := sha256.Sum256(norm)
			h := hex.EncodeToString(sum[:])
			lines[h] = n
			name := path.Base(f)
			byName[name] = append(byName[name], holder{p.Name, h})
		}
	}
	for name, holders := range byName {
		groups := map[string][]string{}
		for _, h := range holders {
			groups[h.hash] = append(groups[h.hash], h.project)
		}
		for hash, members := range groups {
			if len(members) < MinCopies {
				continue
			}
			c := Copy{File: name, Projects: members, Lines: lines[hash]}
			for h2, other := range groups {
				if h2 != hash {
					c.Different = append(c.Different, other...)
				}
			}
			sort.Strings(c.Projects)
			sort.Strings(c.Different)
			r.Copies = append(r.Copies, c)
			r.LinesNow += c.Lines * len(members)
			r.LinesOnce += c.Lines
		}
	}
	sort.Slice(r.Copies, func(i, j int) bool {
		a, b := r.Copies[i], r.Copies[j]
		if sa, sb := a.Lines*(len(a.Projects)-1), b.Lines*(len(b.Projects)-1); sa != sb {
			return sa > sb
		}
		return a.File < b.File
	})
	return r
}

// normalise makes line ends LF and drops trailing white space, and counts the
// lines that are not blank.
func normalise(data []byte) ([]byte, int) {
	var out bytes.Buffer
	n := 0
	for _, l := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		l = strings.TrimRight(l, " \t")
		if l == "" {
			continue
		}
		out.WriteString(l + "\n")
		n++
	}
	return out.Bytes(), n
}
