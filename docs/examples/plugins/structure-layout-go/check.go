package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

// junkDrawers are folder names that say nothing about what is inside.
var junkDrawers = []string{"util", "utils", "helper", "helpers", "common", "shared", "misc"}

// skipped folders are build output and tool state, not the team's layout.
var skipped = []string{"dist", "node_modules", ".git"}

// findJunkDrawers lists the folders below dir whose name is a junk drawer,
// as slash-separated paths relative to dir.
func findJunkDrawers(dir string) ([]string, error) {
	var found []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() || path == dir {
			return err
		}
		if slices.Contains(skipped, d.Name()) {
			return filepath.SkipDir
		}
		if slices.Contains(junkDrawers, strings.ToLower(d.Name())) {
			rel, _ := filepath.Rel(dir, path)
			found = append(found, filepath.ToSlash(rel))
		}
		return nil
	})
	return found, err
}

// runCheck is the layout-check task's command line: `layout-check` in the
// project's directory. It fails, naming each junk drawer.
func runCheck(args []string) error {
	if len(args) != 1 || args[0] != "layout-check" {
		return fmt.Errorf("usage: smith-plugin-widget-layout layout-check (smith runs this program as a plugin)")
	}
	found, err := findJunkDrawers(".")
	if err != nil {
		return err
	}
	if len(found) > 0 {
		return fmt.Errorf("folders named for nothing: %s (name them for what they do)", strings.Join(found, ", "))
	}
	fmt.Println("no junk-drawer folders")
	return nil
}
