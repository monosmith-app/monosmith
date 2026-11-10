package prompt

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// ChangedFiles lists the files of the working tree that differ from HEAD,
// untracked ones included, as workspace-relative paths with forward slashes.
// It runs one git command (two when the workspace is a folder of a larger
// repository). It returns nothing when git fails: a prompt never shows an error.
func ChangedFiles(root string) []string {
	git := gitclient.New(root)
	out, err := git.Run("-c", "core.quotepath=off", "status", "--porcelain=v1", "-z", "--no-renames", "--untracked-files=all")
	if err != nil || out == "" {
		return nil
	}
	prefix := ""
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		// git names paths from the top of the repository.
		if p, err := git.Prefix(); err == nil && p != "" {
			prefix = strings.TrimSuffix(filepath.ToSlash(p), "/") + "/"
		}
	}
	var files []string
	for _, entry := range strings.Split(out, "\x00") {
		if len(entry) < 4 {
			continue
		}
		if f := entry[3:]; strings.HasPrefix(f, prefix) {
			files = append(files, strings.TrimPrefix(f, prefix))
		}
	}
	return files
}
