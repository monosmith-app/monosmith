// Command smith-plugin-widget-lint is a smith toolchain plugin: it fills the
// "lint" slot of widget projects. Started by smith it serves the plugin
// protocol; started by the lint task it contributes, it is the linter.
package main

import (
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

func main() {
	if os.Getenv("SMITH_PLUGIN") == "1" {
		serve.Main(&lint{})
		return
	}
	if err := runLint(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
