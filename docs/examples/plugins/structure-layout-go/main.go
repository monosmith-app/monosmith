// Command smith-plugin-widget-layout is a smith structure plugin: it adds a
// layout-check task to widget projects and makes their lint wait for it.
// Started by smith it serves the plugin protocol; started by the task it
// contributes, it checks the folder it runs in.
package main

import (
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

func main() {
	if os.Getenv("SMITH_PLUGIN") == "1" {
		serve.Main(&layout{})
		return
	}
	if err := runCheck(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
