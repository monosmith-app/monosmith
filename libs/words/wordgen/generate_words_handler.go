// Command wordgen writes the text files the editor extensions read, from the
// catalogue in libs/words/catalog and the key lists in libs/words/surface (ADR
// 0089): package.nls*.json and a typed module for VS Code, messages*.properties
// and a Kotlin object for JetBrains. Run it from libs/words after a change to the
// catalogue or to a surface manifest:
//
//	go run ./wordgen
//
// With -check it writes nothing and fails when a file is not what it would write.
// A Go test runs the same check, so `smith run test` fails on a stale file.
// Every output folder has an argument; the defaults are relative to -root.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	check := flag.Bool("check", false, "write nothing; fail when a file is not up to date")
	root := flag.String("root", "../..", "the repository root, relative to the directory the command runs in")
	wordsDir := flag.String("words", "libs/words", "the library folder with catalog/<locale>/ and surface/, relative to -root")
	surfaces := flag.String("surface", "vscode,jetbrains", "the surfaces to write, separated by commas")
	vscodeModule := flag.String("vscode-module", "apps/vscode/src/words", "folder of the TypeScript module, relative to -root")
	vscodeNls := flag.String("vscode-nls", "apps/vscode", "folder that gets package.nls*.json, relative to -root")
	jbResources := flag.String("jetbrains-resources", "apps/jetbrains/src/main/resources", "folder that gets messages*.properties, relative to -root")
	jbKotlin := flag.String("jetbrains-kotlin", "apps/jetbrains/src/main/kotlin/dev/monosmith/jetbrains/words", "folder of the Kotlin object, relative to -root")
	flag.Parse()
	under := func(p string) string {
		if filepath.IsAbs(p) {
			return p
		}
		return filepath.Join(*root, filepath.FromSlash(p))
	}
	stale, err := Sync(Options{
		Words:              under(*wordsDir),
		VSCodeModule:       under(*vscodeModule),
		VSCodeNls:          under(*vscodeNls),
		JetBrainsResources: under(*jbResources),
		JetBrainsKotlin:    under(*jbKotlin),
		Surfaces:           strings.Split(*surfaces, ","),
		Check:              *check,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "wordgen:", err)
		os.Exit(1)
	}
	for _, f := range stale {
		fmt.Println(f)
	}
	if *check && len(stale) > 0 {
		fmt.Fprintln(os.Stderr, "wordgen: the files above are not up to date: run `go run ./wordgen` in libs/words")
		os.Exit(1)
	}
}
