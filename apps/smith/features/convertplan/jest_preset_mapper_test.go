package convertplan

import (
	"strings"
	"testing"
)

// A workspace's preset extends Nx's with its own setup file; converting must
// keep what it added (a lost setup file made every test connect to telemetry).
func TestRewriteNxJestPresetKeepsWhatTheWorkspaceAdded(t *testing.T) {
	src := `const nxPreset = require('@nx/jest/preset').default

module.exports = {
  ...nxPreset,
  setupFilesAfterEnv:      [require.resolve('./jest.setup.cjs')],
  transformIgnorePatterns: [
    ...(nxPreset.transformIgnorePatterns ?? ['/node_modules/']),
    '/jest-setup-file$',
  ],
}
`
	out, kept := RewriteNxJestPreset(src)
	if !kept {
		t.Fatal("not rewritten")
	}
	for _, want := range []string{"const nxPreset = require('@monosmith/jest-preset')", "...nxPreset,", "setupFilesAfterEnv:      [require.resolve('./jest.setup.cjs')]", "nxPreset.transformIgnorePatterns ?? ['/node_modules/']"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(stripLineComments(out), "@nx/") {
		t.Errorf("code still names Nx:\n%s", out)
	}
}

func TestRewriteNxJestPresetGivesUpOnWhatItCannotRead(t *testing.T) {
	for name, src := range map[string]string{
		"no binding": "module.exports = require('@nx/jest/preset').default\n",
		"another nx": "const nxPreset = require('@nx/jest/preset').default\nmodule.exports = { ...nxPreset, resolver: require.resolve('@nx/jest/plugins/resolver') }\n",
		"esm import": "import nxPreset from '@nx/jest/preset'\nexport default { ...nxPreset }\n",
		"not nx":     "const x = require('./other')\n",
	} {
		if out, kept := RewriteNxJestPreset(src); kept {
			t.Errorf("%s: rewritten to\n%s", name, out)
		}
	}
}
