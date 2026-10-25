// Command smith-plugin-widget-service is a smith scaffold plugin: it writes
// the starter files of a widget service, an app that listens on a port.
// It has no command line of its own; smith starts it as a plugin.
package main

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"

func main() { serve.Main(&scaffold{}) }
