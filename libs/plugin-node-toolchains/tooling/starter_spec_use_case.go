package tooling

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const starterUseCase = "src/greeting/greet.use-case.ts"

// StarterSpec returns the spec of the greeting slice the node plugin starts
// every project with, importing the test API from module ("@jest/globals",
// "vitest"). It takes the suffix the workspace's tests already use
// (greet.use-case.test.ts or .spec.ts), so a new project never mixes the two.
// A project without that slice gets no spec.
func StarterSpec(workspaceRoot string, params protocol.GenerateParams, module string) map[string]string {
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory), filepath.FromSlash(starterUseCase))); err != nil {
		return nil
	}
	name := "src/greeting/greet.use-case." + scaffolding.TestSuffix(workspaceRoot) + ".ts"
	return map[string]string{name: `import { describe, expect, it } from '` + module + `'
import { greet } from './greet.use-case.ts'

describe('greet', () => {
  it('greets by name', () => {
    expect(greet('world').message).toBe('Hello, world!')
  })
})
`}
}
