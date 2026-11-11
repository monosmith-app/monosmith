package discovery

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// CrateBinaries names the programs `cargo build -p <crate>` writes into the
// profile folder: each [[bin]] table, then, unless autobins is false,
// src/main.rs under the crate's name and each src/bin/<name>.rs under its
// file name. A [[bin]] that names the same file as a default takes its place
// (crate loomas-cli with [[bin]] name = "loomas", path = "src/main.rs" builds
// only loomas). A program in src/bin/<name>/main.rs is left out: smith
// declares such a folder as a project of its own.
//
// Parameters: dir is the crate's folder on disk; c the parsed Cargo.toml.
// It returns the program names in the order cargo lists them, without
// duplicates; none for a library.
func CrateBinaries(dir string, c Crate) []string {
	var names []string
	seenName, seenPath := map[string]bool{}, map[string]bool{}
	add := func(name string) {
		if name != "" && !seenName[name] {
			seenName[name] = true
			names = append(names, name)
		}
	}
	for _, b := range c.Bins {
		add(b.Name)
		switch {
		case b.Path != "":
			seenPath[path.Clean(filepath.ToSlash(b.Path))] = true
		case b.Name == c.Name:
			seenPath["src/main.rs"] = true
		default:
			seenPath["src/bin/"+b.Name+".rs"] = true
		}
	}
	if !c.AutoBins {
		return names
	}
	if isFile(filepath.Join(dir, "src", "main.rs")) && !seenPath["src/main.rs"] {
		add(c.Name)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "src", "bin"))
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".rs") && !seenPath["src/bin/"+e.Name()] {
			files = append(files, strings.TrimSuffix(e.Name(), ".rs"))
		}
	}
	sort.Strings(files)
	for _, f := range files {
		add(f)
	}
	return names
}

// isFile is true when p names a regular file.
func isFile(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
