package convertplan

import "testing"

func TestRollupDelegateFindsTheSharedBuildAOneLineConfigRequires(t *testing.T) {
	files := []string{"rollup.base.cjs", "packages/a/rollup.config.cjs", "packages/b/rollup.config.cjs", "packages/c/rollup.config.cjs", "other.cjs"}
	contents := map[string]string{
		"rollup.base.cjs":              "const { withNx } = require('@nx/rollup/with-nx')\nmodule.exports = () => withNx({ main: './src/index.ts' })\n",
		"packages/a/rollup.config.cjs": "// Nx infers the build target from this file (@nx/rollup/plugin).\nmodule.exports = require('../../rollup.base.cjs')()\n",
		"packages/b/rollup.config.cjs": "const { withNx } = require('@nx/rollup/with-nx')\nmodule.exports = withNx({ main: './src/index.ts' })\n",
		"packages/c/rollup.config.cjs": "module.exports = require('../../other.cjs')()\n",
		"other.cjs":                    "module.exports = () => ({})\n",
	}
	read := func(f string) string { return contents[f] }
	file, text, ok := rollupDelegate("packages/a/rollup.config.cjs", files, read)
	if !ok || file != "rollup.base.cjs" || text != contents["rollup.base.cjs"] {
		t.Errorf("a: %q %v", file, ok)
	}
	if _, _, ok := rollupDelegate("packages/b/rollup.config.cjs", files, read); ok {
		t.Error("a config that is a build itself delegates to nothing")
	}
	if _, _, ok := rollupDelegate("packages/c/rollup.config.cjs", files, read); ok {
		t.Error("a require of a file that is not a withNx build is not a delegation")
	}
}

// A comment that names Nx gives no import to report: the fallback answers "nx".
func TestANxNameInACommentIsNotAnImport(t *testing.T) {
	text := "// Nx infers this build from the file (@nx/rollup/plugin globs it).\nmodule.exports = require('../../base.cjs')()\n"
	if got := nxImport(text); got != "nx" {
		t.Errorf("nxImport = %q", got)
	}
	if got := nxImport("const x = require('@nx/devkit') // not @nx/other"); got != "@nx/devkit" {
		t.Errorf("nxImport = %q", got)
	}
}
