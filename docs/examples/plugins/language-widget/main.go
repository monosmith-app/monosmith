// Command smith-plugin-widget is a smith language plugin for "widget"
// projects (a directory with a widget.json). Started by smith it serves the
// plugin protocol on stdin and stdout; started by a person, or by one of the
// tasks it contributes, it is a small command line tool.
package main

import (
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

func main() {
	// smith sets SMITH_PLUGIN=1 in the environment of every plugin it starts.
	if os.Getenv("SMITH_PLUGIN") == "1" {
		serve.Main(&widget{})
		return
	}
	if err := runTool(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
