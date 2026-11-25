package discovery

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestApplyFunctionPackage(t *testing.T) {
	scripts := map[string]string{"build": "tsc -p tsconfig.build.json && smith func assemble", "package": "smith func package --with-dependencies"}
	var tasks map[string]protocol.Task
	beforeEach := func() {
		tasks = map[string]protocol.Task{"package": TaskFor("package", "apps/api", scripts)}
		ApplyFunctionPackage(tasks, "apps/api", scripts)
	}

	t.Run("package is cached after the builds, and does not build again", func(t *testing.T) {
		beforeEach()
		p := tasks["package"]
		if !p.Cache || !slices.Equal(p.DependsOn, []string{"build", "^build"}) || p.Env[PrebuiltEnv] != "1" || p.Verify {
			t.Fatalf("package %+v", p)
		}
		if !slices.Equal(p.Outputs, []string{"../../dist/apps/api", "../../dist/drop/api"}) {
			t.Fatalf("the zip and its dist/drop copy are the outputs: %v", p.Outputs)
		}
	})

	t.Run("package-check runs in verify", func(t *testing.T) {
		beforeEach()
		c := tasks[PackageCheckTask]
		if !c.Verify || !c.Cache || !slices.Equal(c.Command, []string{"{smith}", "func", "package", "--check"}) || c.Env[PrebuiltEnv] != "1" {
			t.Fatalf("package-check %+v", c)
		}
	})

	t.Run("only smith func package is a function package", func(t *testing.T) {
		if !FunctionPackage(" smith func package") || FunctionPackage("npm pack") || FunctionPackage("npm run build && cd x && npx electron-forge package") {
			t.Fatal("FunctionPackage")
		}
	})
}
