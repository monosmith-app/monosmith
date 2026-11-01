package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// lockfilePlatformCheck is the name of the check that package-lock.json holds
// the packages every platform needs.
const lockfilePlatformCheck = "package-lock.json holds the packages every platform needs"

// platformRemedy is the remedy for a lockfile that lacks packages: regenerate
// it, and if npm ci still fails list the packages in the root devDependencies.
// An entry of package.json overrides for the same package must then be
// "$<name>", or npm stops with "Override ... conflicts with direct dependency".
func platformRemedy(shown []string) string {
	return "regenerate package-lock.json and commit it (`smith doctor --fix` does it: it removes the file and runs npm install --package-lock-only)\n" +
		"if npm ci still fails on Linux, add the missing packages to the root devDependencies (npm install -D " + strings.Join(shown, " ") + ")\n" +
		"if the root package.json \"overrides\" already names one of them, keep it in devDependencies too and set the override to \"$<name>\" (for example \"@emnapi/core\": \"$@emnapi/core\"): npm refuses an override that differs from a direct dependency"
}

// lockfilePlatformChecks reads package-lock.json in the workspace root.
func lockfilePlatformChecks(s *session.Session) []protocol.Finding {
	return checkLockfilePlatformDeps(s.Root)
}

// checkLockfilePlatformDeps reports a package-lock.json that leaves out
// packages its entries depend on, typically the optional packages (@emnapi/*,
// platform binaries) of the OSes the lockfile was not made on. It says nothing
// when there is no readable lockfile.
func checkLockfilePlatformDeps(root string) []protocol.Finding {
	raw, err := os.ReadFile(filepath.Join(root, "package-lock.json"))
	if err != nil {
		return nil
	}
	missing, err := missingLockfileDeps(raw)
	if err != nil {
		return nil
	}
	f := protocol.Finding{Check: lockfilePlatformCheck, OK: len(missing) == 0, Warning: true}
	if f.OK {
		return []protocol.Finding{f}
	}
	shown := missing
	if len(shown) > 5 {
		shown = shown[:5]
	}
	f.Detail = fmt.Sprintf("npm ci on another OS (CI, Linux) fails: the lockfile lacks %d package(s) it depends on, such as %s. npm leaves out the optional packages of other platforms (such as the wasm fallbacks @emnapi/core and @emnapi/runtime) when the lockfile is made on one OS", len(missing), strings.Join(shown, ", "))
	f.Remedy = platformRemedy(shown)
	f.Fix = &protocol.Fix{
		Description: "regenerate package-lock.json so it holds the packages of every platform",
		Steps: []protocol.FixStep{
			{Description: "remove package-lock.json", Argv: []string{"node", "-e", "require('fs').rmSync('package-lock.json',{force:true})"}},
			{Description: "resolve every platform's packages into a new package-lock.json", Argv: []string{"npm", "install", "--package-lock-only", "--no-audit", "--no-fund"}},
		},
	}
	return []protocol.Finding{f}
}
