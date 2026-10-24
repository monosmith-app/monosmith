package pipeline

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// configNames are the root config files whose imports must be installed for
// a tool to start.
var configNames = []string{"eslint.config.js", "eslint.config.mjs", "eslint.config.cjs", "eslint.config.ts", "eslint.config.mts"}

var configImport = regexp.MustCompile(`(?:from\s+|import\s*\(\s*|require\s*\(\s*|import\s+)['"]([^'"]+)['"]`)

var nodeBuiltins = map[string]bool{
	"assert": true, "buffer": true, "child_process": true, "cluster": true, "console": true, "constants": true,
	"crypto": true, "dgram": true, "dns": true, "events": true, "fs": true, "http": true, "http2": true, "https": true,
	"inspector": true, "module": true, "net": true, "os": true, "path": true, "perf_hooks": true, "process": true,
	"punycode": true, "querystring": true, "readline": true, "repl": true, "stream": true, "string_decoder": true,
	"timers": true, "tls": true, "tty": true, "url": true, "util": true, "v8": true, "vm": true, "worker_threads": true, "zlib": true,
}

// CheckConfigImports checks that each package the workspace root's ESLint
// config imports is installed, since the tool cannot start without it. It
// judges only once dependencies are installed.
func CheckConfigImports(workspaceRoot string) (protocol.Finding, bool) {
	if _, err := os.Stat(filepath.Join(workspaceRoot, "node_modules")); err != nil {
		return protocol.Finding{}, false
	}
	var problems, missing []string
	found := false
	for _, name := range configNames {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, name))
		if err != nil {
			continue
		}
		found = true
		seen := map[string]bool{}
		for _, m := range configImport.FindAllStringSubmatch(string(raw), -1) {
			pkg := packageOf(m[1])
			if pkg == "" || seen[pkg] {
				continue
			}
			seen[pkg] = true
			if _, err := os.Stat(filepath.Join(workspaceRoot, "node_modules", filepath.FromSlash(pkg))); err != nil {
				problems = append(problems, name+" imports "+pkg+", which is not installed")
				missing = append(missing, pkg)
			}
		}
	}
	if !found {
		return protocol.Finding{}, false
	}
	sort.Strings(problems)
	finding := protocol.Finding{
		Check:  "the packages the ESLint config imports are installed",
		OK:     len(problems) == 0,
		Detail: strings.Join(problems, "; "),
		Remedy: "install the package (npm i -D <package> at the workspace root), or change the import to one that is installed",
	}
	if !finding.OK {
		finding.Fix = manifestfixes.InstallMissingPackagesFix(workspaceRoot, missing)
	}
	return finding, true
}

// packageOf is the package a bare import names, "" for a relative path, a
// Node built-in or a node: URL.
func packageOf(spec string) string {
	if spec == "" || strings.HasPrefix(spec, ".") || strings.HasPrefix(spec, "/") || strings.HasPrefix(spec, "node:") || strings.Contains(spec, ":") {
		return ""
	}
	parts := strings.Split(spec, "/")
	if strings.HasPrefix(spec, "@") {
		if len(parts) < 2 {
			return ""
		}
		return parts[0] + "/" + parts[1]
	}
	if nodeBuiltins[parts[0]] {
		return ""
	}
	return parts[0]
}
