package release

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// journalPath is a project's journal, under the directory its release
// assets and packages are built in.
func journalPath(root, project string) string {
	return filepath.Join(root, "dist", "release", project, "journal.json")
}

// ReadJournal is a project's journal; a project that never released has an
// empty one.
func ReadJournal(root, project string) (Journal, error) {
	j := Journal{Project: project}
	raw, err := os.ReadFile(journalPath(root, project))
	if errors.Is(err, fs.ErrNotExist) {
		return j, nil
	}
	if err != nil {
		return j, err
	}
	if err := json.Unmarshal(raw, &j); err != nil {
		return j, err
	}
	return j, nil
}

// AppendJournal adds entries to a project's journal. Nothing is ever
// rewritten or removed.
func AppendJournal(root, project string, entries []JournalEntry) error {
	if len(entries) == 0 {
		return nil
	}
	j, err := ReadJournal(root, project)
	if err != nil {
		return err
	}
	j.Entries = append(j.Entries, entries...)
	raw, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}
	path := journalPath(root, project)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

// ReadJournals reads every project's journal, by project name.
func ReadJournals(root string) ([]Journal, error) {
	dirs, err := os.ReadDir(filepath.Join(root, "dist", "release"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Journal
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		j, err := ReadJournal(root, d.Name())
		if err != nil {
			return nil, err
		}
		if len(j.Entries) > 0 {
			out = append(out, j)
		}
	}
	sort.Slice(out, func(i, k int) bool { return out[i].Project < out[k].Project })
	return out, nil
}

// JournalHas reports whether a project's journal holds a published or
// existing package whose name starts with name (a wheel, gem or .nupkg file
// name starts with its package name), at a version.
func JournalHas(j Journal, name, version string) bool {
	want := normalizeName(name)
	for _, e := range j.Entries {
		if e.Version == version && e.Step == "package" && strings.HasPrefix(normalizeName(e.Target), want) {
			return true
		}
	}
	return false
}

func normalizeName(s string) string {
	return strings.ToLower(strings.NewReplacer("_", "-", ".", "-").Replace(s))
}
