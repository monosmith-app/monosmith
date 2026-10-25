// Command smith-plugin-widget-bundle is a smith deploy plugin: it adds a
// bundle task to widget apps and, on release, deploys the bundle to a
// directory. Started by smith it serves the plugin protocol; started by the
// bundle task it contributes, it makes the bundle.
package main

import (
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

func main() {
	if os.Getenv("SMITH_PLUGIN") == "1" {
		serve.Main(&bundle{})
		return
	}
	if err := runBundle(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runBundle is the bundle task's command line:
//
//	smith-plugin-widget-bundle bundle <build-dir> <out-dir>
//
// It reads <build-dir>/widget.txt, which the app's build wrote, and writes
// <out-dir>/bundle.txt.
func runBundle(args []string) error {
	if len(args) != 3 || args[0] != "bundle" {
		return fmt.Errorf("usage: smith-plugin-widget-bundle bundle <build-dir> <out-dir> (smith runs this program as a plugin)")
	}
	built, err := os.ReadFile(args[1] + "/widget.txt")
	if err != nil {
		return fmt.Errorf("nothing to bundle: %w (run the app's build first)", err)
	}
	if err := os.MkdirAll(args[2], 0o755); err != nil {
		return err
	}
	return os.WriteFile(args[2]+"/bundle.txt", bundleOf(built), 0o644)
}
