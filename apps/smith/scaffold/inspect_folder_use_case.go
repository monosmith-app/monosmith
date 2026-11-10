package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// maxInspected caps the files InspectFolder lists, so a huge folder does not stall smith new.
const maxInspected = 5000

// FolderContent is what a folder holds before smith writes into it.
type FolderContent struct {
	// Files are slash-separated paths relative to the folder, sorted. A
	// node_modules folder counts as one entry and is not walked; the entries of
	// ignoredInEmptyFolder (.git, .vscode, OS files) are skipped.
	Files []string
	// Secrets are the Files whose names look like secrets (IsSecretName).
	Secrets []string
}

// InspectFolder lists what a folder holds. A folder that does not exist holds
// nothing. It reads names only.
//
// Parameters: dir is the folder to read.
// Returns: the content; an error only when the folder exists and cannot be read.
func InspectFolder(dir string) (FolderContent, error) {
	var c FolderContent
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return c, err
	}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == dir {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		rel = filepath.ToSlash(rel)
		if ignoredInEmptyFolder[d.Name()] {
			// The repository, editor settings and OS files do not count as content.
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if d.Name() == "node_modules" {
				c.Files = append(c.Files, rel+"/")
				return filepath.SkipDir
			}
			return nil
		}
		if len(c.Files) >= maxInspected {
			return filepath.SkipAll
		}
		c.Files = append(c.Files, rel)
		if IsSecretName(rel) {
			c.Secrets = append(c.Secrets, rel)
		}
		return nil
	})
	sort.Strings(c.Files)
	sort.Strings(c.Secrets)
	return c, err
}
