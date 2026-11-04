// Command flaggen writes the files generated from the flag vocabulary
// (libs/actions/flagvocabulary/flag_vocabulary.json): the flag constants of the
// VS Code extension and the JetBrains plugin and the reference table
// docs/reference/flag-vocabulary.md. Run it from libs/actions after a change to the
// vocabulary:
//
//	go run ./flaggen
//
// With -check it writes nothing and fails when a file is not what it would write.
// A Go test runs the same check, so `smith run test` fails on a stale file.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/libs/actions/flagclients"
)

func main() {
	check := flag.Bool("check", false, "write nothing; fail when a file is not up to date")
	root := flag.String("root", "../..", "the repository root, relative to the directory the command runs in")
	flag.Parse()
	stale, err := flagclients.Sync(*root, *check)
	if err != nil {
		fmt.Fprintln(os.Stderr, "flaggen:", err)
		os.Exit(1)
	}
	for _, f := range stale {
		fmt.Println(f)
	}
	if *check && len(stale) > 0 {
		fmt.Fprintln(os.Stderr, "flaggen: the files above are not up to date: run `go run ./flaggen` in libs/actions")
		os.Exit(1)
	}
}
