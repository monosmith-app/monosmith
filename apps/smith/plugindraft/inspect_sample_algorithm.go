// Package plugindraft looks at a sample project in a language smith does not
// know and drafts what a language plugin for it needs to say: the manifest
// files, the version line, the tasks and the publish command (`smith plugin
// draft`, ADR 0061). The draft is a description for the person writing the
// plugin, not a plugin: smith has no declarative plugin format.
package plugindraft

import (
	"encoding/json"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
)

// Manifest is a file at the sample's root that names the project and gives it
// a version.
type Manifest struct {
	File string `json:"file"`
	// NameLine and VersionLine are the lines that hold the name and the
	// version, trimmed; the plugin's version/write rewrites VersionLine.
	NameLine    string `json:"nameLine"`
	VersionLine string `json:"versionLine"`
}

// Task is a command the sample already runs for a task name smith knows.
type Task struct {
	Name    string   `json:"name"`
	Command []string `json:"command"`
	// Source is where it was found ("Makefile", "package.json scripts").
	Source string `json:"source"`
}

// Draft is what Inspect found.
type Draft struct {
	Manifests []Manifest `json:"manifests"`
	Tasks     []Task     `json:"tasks"`
	// Publish are lines of CI files and scripts that publish, with their file.
	Publish []string `json:"publish"`
}

var (
	nameLine    = regexp.MustCompile(`(?i)(^\s*\.?"?(name|artifactId|PackageId)"?\s*[:=])|(<(name|artifactId|PackageId)>)`)
	versionLine = regexp.MustCompile(`(?i)(^\s*\.?"?version"?\s*[:=])|(<version>)`)
	target      = regexp.MustCompile(`^([a-zA-Z][\w.-]*)\s*:($|[^=])`)
	publishLine = regexp.MustCompile(`(?i)\b(publish|upload|push)\b`)
	taskNames   = map[string]bool{"build": true, "test": true, "lint": true, "format": true, "fmt": true, "check": true, "clean": true, "dev": true, "run": true, "typecheck": true, "publish": true, "release": true}
	skipFiles   = regexp.MustCompile(`(?i)^(readme|license|changelog|contributing|\.).*|.*\.(lock|sum|md|txt)$|^package-lock\.json$`)
)

const maxBytes = 64 * 1024

// Inspect reads the files at the root of sample (and its CI files) and drafts
// what it can say about the language's project.
//
// Parameters: sample is the folder of a small working project.
// Returns the draft; it never fails on a file it cannot read, it skips it.
func Inspect(sample fs.FS) Draft {
	var d Draft
	entries, err := fs.ReadDir(sample, ".")
	if err != nil {
		return d
	}
	for _, e := range entries {
		if e.IsDir() || skipFiles.MatchString(e.Name()) {
			continue
		}
		text, ok := read(sample, e.Name())
		if !ok {
			continue
		}
		if m, ok := manifestOf(e.Name(), text); ok {
			d.Manifests = append(d.Manifests, m)
		}
		d.Tasks = append(d.Tasks, scriptTasks(e.Name(), text)...)
		d.Tasks = append(d.Tasks, targetTasks(e.Name(), text)...)
		d.Publish = append(d.Publish, publishLines(e.Name(), text)...)
	}
	for _, dir := range []string{".github/workflows"} {
		files, _ := fs.ReadDir(sample, dir)
		for _, f := range files {
			if text, ok := read(sample, path.Join(dir, f.Name())); ok {
				d.Publish = append(d.Publish, publishLines(path.Join(dir, f.Name()), text)...)
			}
		}
	}
	sort.Slice(d.Tasks, func(i, j int) bool {
		return d.Tasks[i].Name < d.Tasks[j].Name || (d.Tasks[i].Name == d.Tasks[j].Name && d.Tasks[i].Source < d.Tasks[j].Source)
	})
	return d
}

func read(fsys fs.FS, name string) (string, bool) {
	b, err := fs.ReadFile(fsys, name)
	if err != nil || len(b) > maxBytes || strings.IndexByte(string(b), 0) >= 0 {
		return "", false
	}
	return string(b), true
}

func manifestOf(file, text string) (Manifest, bool) {
	m := Manifest{File: file}
	for _, l := range strings.Split(text, "\n") {
		switch {
		case m.NameLine == "" && nameLine.MatchString(l):
			m.NameLine = strings.TrimSpace(l)
		case m.VersionLine == "" && versionLine.MatchString(l):
			m.VersionLine = strings.TrimSpace(l)
		}
	}
	return m, m.NameLine != "" && m.VersionLine != ""
}

// scriptTasks reads the "scripts" object of a JSON file.
func scriptTasks(file, text string) []Task {
	if !strings.HasSuffix(file, ".json") {
		return nil
	}
	var doc struct{ Scripts map[string]string }
	if json.Unmarshal([]byte(text), &doc) != nil {
		return nil
	}
	var out []Task
	for name, cmd := range doc.Scripts {
		if taskNames[name] {
			out = append(out, Task{Name: name, Command: strings.Fields(cmd), Source: file + " scripts"})
		}
	}
	return out
}

// targetTasks reads the targets of a Makefile or a justfile.
func targetTasks(file, text string) []Task {
	var tool string
	switch strings.ToLower(file) {
	case "makefile", "gnumakefile":
		tool = "make"
	case "justfile":
		tool = "just"
	default:
		return nil
	}
	var out []Task
	for _, l := range strings.Split(text, "\n") {
		if m := target.FindStringSubmatch(l); m != nil && taskNames[m[1]] {
			out = append(out, Task{Name: m[1], Command: []string{tool, m[1]}, Source: file})
		}
	}
	return out
}

func publishLines(file, text string) []string {
	var out []string
	for _, l := range strings.Split(text, "\n") {
		t := strings.TrimSpace(l)
		if publishLine.MatchString(t) && (strings.HasPrefix(t, "run:") || strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "\"publish") || strings.HasPrefix(t, "publish")) {
			out = append(out, file+": "+t)
		}
	}
	return out
}
