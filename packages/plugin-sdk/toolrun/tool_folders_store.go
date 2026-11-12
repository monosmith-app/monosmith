package toolrun

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

// ToolFoldersFile is the file where smith records, per user, the folders of
// tools that are installed but not on PATH (a winget portable package), after
// the person chose to let smith use them from there. It sits beside smith's
// other per-user files.
//
// Parameters: home is smith's per-user directory (SMITH_HOME or ~/.smith).
// Returns: the file path, or "" when home is "".
func ToolFoldersFile(home string) string {
	if home == "" {
		return ""
	}
	return filepath.Join(home, "tool-folders.json")
}

// ReadToolFolders reads the recorded folders by tool name.
//
// Parameters: file is ToolFoldersFile.
// Returns: the folders (empty when the file does not exist), or the read or
// parse error.
func ReadToolFolders(file string) (map[string]string, error) {
	folders := map[string]string{}
	if file == "" {
		return folders, nil
	}
	data, err := os.ReadFile(file)
	if errors.Is(err, os.ErrNotExist) {
		return folders, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &folders); err != nil {
		return nil, err
	}
	return folders, nil
}

// RecordToolFolder records the folder that holds a tool's program, so every
// later smith command puts it on PATH (UseToolFolders).
//
// Parameters: file is ToolFoldersFile; tool is the tool's name; dir is the
// absolute folder.
// Returns: an error when the file cannot be read or written.
func RecordToolFolder(file, tool, dir string) error {
	if file == "" {
		return errors.New("smith has no folder for its own files on this machine (set SMITH_HOME)")
	}
	folders, err := ReadToolFolders(file)
	if err != nil {
		return err
	}
	folders[tool] = dir
	data, err := json.MarshalIndent(folders, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	return os.WriteFile(file, append(data, '\n'), 0o644)
}

// WithToolFolders adds the recorded folders at the end of a PATH value. A
// folder that PATH already holds is not added again, so a tool the person put
// on PATH keeps its place, and a folder never shadows a program PATH finds
// first. The folders are added in the order of the tool names.
//
// Parameters: path is the PATH value; folders are the recorded folders by tool.
// Returns: the new PATH value.
func WithToolFolders(path string, folders map[string]string) string {
	names := make([]string, 0, len(folders))
	for name := range folders {
		names = append(names, name)
	}
	slices.Sort(names)
	entries := filepath.SplitList(path)
	same := func(a, b string) bool {
		a, b = filepath.Clean(a), filepath.Clean(b)
		if runtime.GOOS == "windows" {
			return strings.EqualFold(a, b)
		}
		return a == b
	}
	for _, name := range names {
		dir := folders[name]
		if dir == "" || slices.ContainsFunc(entries, func(e string) bool { return e != "" && same(e, dir) }) {
			continue
		}
		entries = append(entries, dir)
	}
	return strings.Join(entries, string(os.PathListSeparator))
}

// UseToolFolders puts the recorded folders that still exist on the PATH of
// smith's own process, so the tools it detects, the plugins it starts and the
// tasks it runs all find the programs there. The person's own PATH (the user
// and machine settings) is not changed.
//
// Parameters: file is ToolFoldersFile.
// Returns: an error when the file cannot be read; PATH is then unchanged.
func UseToolFolders(file string) error {
	folders, err := ReadToolFolders(file)
	if err != nil || len(folders) == 0 {
		return err
	}
	for name, dir := range folders {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			delete(folders, name)
		}
	}
	return os.Setenv("PATH", WithToolFolders(os.Getenv("PATH"), folders))
}
