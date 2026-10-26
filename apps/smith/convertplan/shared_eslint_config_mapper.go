package convertplan

import (
	"path"
	"strings"
)

// MnciESLintConfig is the shared ESLint config the old mnci installed.
const MnciESLintConfig = "@mnci/eslint-config"

// planSharedESLintConfig tells a workspace whose root ESLint config still
// imports the old shared config that smith's own rules are not active, and
// how to switch. Convert does not change lint behaviour by itself.
func planSharedESLintConfig(plan *Plan, files []string, read func(string) string) {
	for _, f := range files {
		if path.Dir(f) != "." || !strings.HasPrefix(path.Base(f), "eslint.config.") {
			continue
		}
		code := stripLineComments(read(f))
		if !strings.Contains(code, "'"+MnciESLintConfig+"'") && !strings.Contains(code, `"`+MnciESLintConfig+`"`) {
			continue
		}
		plan.Decide("", f, "still uses the old shared lint config "+MnciESLintConfig+", so smith's own lint rules (@monosmith/eslint-config, which also keep code organised by feature) are not active. Lint keeps working as it does today. To switch: npm i -D @monosmith/eslint-config, then make the file two lines (import monosmith from '@monosmith/eslint-config'; export default monosmith({ workspaceRoot: import.meta.dirname })). After that "+MnciESLintConfig+" and the plugins only it needed (@eslint/js, typescript-eslint, ...) can come out of the root devDependencies")
		return
	}
}
