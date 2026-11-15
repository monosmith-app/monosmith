// Command smith is a monorepo tool: it scaffolds apps, libs and packages in any
// language a plugin supports, runs their tasks in dependency order from a
// cache, generates CI pipelines and releases packages from conventional
// commits.
package main

import (
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/features/commandline"
)

// version is stamped at build time with -ldflags "-X main.version=1.2.3".
var version = "dev"

func main() {
	os.Exit(commandline.Execute(version, os.Args[1:]))
}
