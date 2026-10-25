package adoptflow

import (
	"os"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// TestMain lets the test binary stand in for smith: the plugin host starts
// built-in plugins as `<self> plugin serve <name>`, and without this the test
// binary would re-run the whole suite for each plugin and multiply without end.
func TestMain(m *testing.M) {
	if len(os.Args) == 4 && os.Args[1] == "plugin" && os.Args[2] == "serve" {
		ctor, ok := pluginhost.Builtin(os.Args[3])
		if !ok {
			os.Exit(2)
		}
		if err := serve.Serve(ctor("test"), os.Stdin, os.Stdout); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}
