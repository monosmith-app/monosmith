package convertplan

import (
	"regexp"
	"strings"
)

var nxPresetBinding = regexp.MustCompile(`(?m)^([ \t]*)(?:const|let|var)[ \t]+(\w+)[ \t]*=[ \t]*require\(\s*['"]@nx/jest/preset['"]\s*\)(?:\.default)?[ \t]*;?[ \t]*(\r?)$`)

// RewriteNxJestPreset takes Nx out of a jest.preset.js that builds on
// @nx/jest/preset (const nxPreset = require('@nx/jest/preset').default, then
// `...nxPreset` and whatever the workspace adds to it, such as a setup file):
// the binding becomes the preset's values written out, and the rest of the
// file is kept as it is. kept is false when the file has no such binding or
// still names Nx after the rewrite, so the caller must replace it.
func RewriteNxJestPreset(text string) (out string, kept bool) {
	m := nxPresetBinding.FindStringSubmatch(text)
	if m == nil {
		return "", false
	}
	binding := m[1] + "// Written by smith convert: the values @nx/jest/preset provided, without Nx." + m[3] + "\n" +
		m[1] + "const " + m[2] + " = " + jestPresetValues
	out = strings.Replace(text, m[0], binding, 1)
	code := stripLineComments(out)
	if strings.Contains(code, "@nx/") || strings.Contains(code, "@nrwl/") {
		return "", false
	}
	return out, true
}
