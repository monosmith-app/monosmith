package discovery

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// fallbackKind is the kind of a module that neither says its kind nor sits
// under apps/, libs/ or packages/: a module whose top directory holds a
// package main is a program, so an app; any other is a lib.
func fallbackKind(dir string) protocol.ProjectKind {
	if holdsMainPackage(dir) {
		return protocol.ProjectApp
	}
	return protocol.ProjectLib
}

// holdsMainPackage reports whether a non-test .go file directly in dir is in
// package main.
func holdsMainPackage(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(f)
		found := false
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if rest, ok := strings.CutPrefix(line, "package "); ok {
				found = strings.TrimSpace(strings.SplitN(rest, "//", 2)[0]) == "main"
				break
			}
		}
		f.Close()
		if found {
			return true
		}
	}
	return false
}
