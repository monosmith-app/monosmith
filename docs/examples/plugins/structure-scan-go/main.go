// Command smith-plugin-widget-scan is a smith structure plugin whose doctor
// walks the files of widget projects: it stops when smith cancels the request
// and it takes four scans at the same time. It only speaks the plugin protocol.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

func main() {
	if os.Getenv("SMITH_PLUGIN") != "1" {
		fmt.Fprintln(os.Stderr, "smith-plugin-widget-scan is a smith plugin: smith starts it and speaks the plugin protocol on its stdin and stdout")
		os.Exit(1)
	}
	serve.Main(&scan{delay: delayFromEnvironment()})
}

// delayFromEnvironment reads WIDGET_SCAN_DELAY_MS: a pause before each file,
// which stands in for a slow disk so a test can cancel a scan in the middle.
func delayFromEnvironment() time.Duration {
	ms, err := time.ParseDuration(os.Getenv("WIDGET_SCAN_DELAY_MS") + "ms")
	if err != nil || ms < 0 {
		return 0
	}
	return ms
}
