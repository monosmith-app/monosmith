package toolchainversions

// The ranges the Node stacks pin. Nothing installs "latest": bump a range
// here, deliberately, and every stack follows.
const (
	// TypeScript stays on 6.x until ts-jest and typescript-eslint support 7.
	TypeScript = "~6.0.3"
	// TypesNode follows Node 24, the Node that CI installs.
	TypesNode = "^24.19.2"
	// Esbuild is the newest esbuild. The TypeScript and VS Code stacks both
	// use it, so a workspace with both holds one range.
	Esbuild = "^0.28.2"
	// Vitest is the test runner of the vitest stack.
	Vitest = "^5.0.5"
	// Jest is the version of jest and of @jest/globals, which must match.
	Jest = "^30.5.2"
)

// Ranges maps each package named above to its range.
//
// Returns: a new map; the caller may change it.
func Ranges() map[string]string {
	return map[string]string{
		"typescript":    TypeScript,
		"@types/node":   TypesNode,
		"esbuild":       Esbuild,
		"vitest":        Vitest,
		"jest":          Jest,
		"@jest/globals": Jest,
	}
}
