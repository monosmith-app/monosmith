package pipeline

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// PHPVersion is installed when .php-version does not pin one; bump it
// deliberately.
const PHPVersion = "8.4"

// Version is the PHP .php-version pins, else PHPVersion.
func Version(workspaceRoot string) string {
	if raw, err := os.ReadFile(filepath.Join(workspaceRoot, ".php-version")); err == nil {
		if v := strings.TrimSpace(string(raw)); v != "" {
			return v
		}
	}
	return PHPVersion
}

func runtime(version string) protocol.Toolchain {
	return protocol.Toolchain{
		Name: "php", Version: version,
		GitHub: &protocol.ActionStep{Uses: "shivammathur/setup-php@v2", Inputs: []protocol.Input{{Name: "php-version", Value: version}, {Name: "tools", Value: "composer:v2"}}},
		// Elsewhere (Azure's image has no pinned PHP; GitLab's Debian image
		// none): the distribution's PHP and Composer's installer.
		Install: "SUDO=$(command -v sudo || true)\n" +
			"$SUDO apt-get update -qq && $SUDO apt-get install -y -qq php-cli php-xml php-mbstring php-curl php-zip unzip\n" +
			"curl -fsSL https://getcomposer.org/installer | $SUDO php -- --install-dir=/usr/local/bin --filename=composer",
		Devcontainer: &protocol.DevcontainerFeature{ID: "ghcr.io/devcontainers/features/php:1", Options: map[string]string{"version": version, "installComposer": "true"}},
	}
}

// CI returns the PHP contribution: PHP with Composer and one composer
// install for the workspace. Packagist needs no secret: it reads tags.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	cacheFiles := []string{"composer.json", "composer.lock"}
	for _, p := range params.Projects {
		cacheFiles = append(cacheFiles, path.Join(p.Root, "composer.json"))
	}
	return protocol.CIResult{
		Toolchains: []protocol.Toolchain{runtime(Version(workspaceRoot))},
		Stores:     []protocol.PackageStore{{Name: "composer", Path: "~/.cache/composer", Env: "COMPOSER_CACHE_DIR", Files: []string{"**/composer.lock"}}},
		Setup:      []protocol.Step{{Name: "Install the Composer packages", Command: []string{"composer", "install", "--no-interaction", "--no-progress"}, Blocking: true}},
		// Known-vulnerable packages from Composer's own audit (Composer 2.4
		// and later), reported (like pip-audit) and re-checked only when a
		// composer.json or the lockfile changes.
		Audit: []protocol.Step{{Name: "composer audit", Command: []string{"sh", "-c", auditScript}, Blocking: false, CacheFiles: cacheFiles}},
	}
}

// auditScript runs composer audit, or says what is missing instead of
// failing with "command not found". Composer exits non-zero when it finds
// advisories.
const auditScript = "if composer --version >/dev/null 2>&1; then composer audit --no-interaction; " +
	"else echo 'composer is not installed: install it (https://getcomposer.org/download) to check packages for known vulnerabilities' >&2; exit 1; fi"

// probe runs a tool for its output; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// Doctor checks that PHP and Composer are installed.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	_, phpErr := probe(workspaceRoot, []string{"php", "--version"})
	_, composerErr := probe(workspaceRoot, []string{"composer", "--version"})
	return protocol.DoctorResult{Findings: []protocol.Finding{
		{Check: "PHP is installed", OK: phpErr == nil, Detail: "php is not on PATH", Remedy: "install PHP " + Version(workspaceRoot) + " (https://www.php.net/downloads)"},
		{Check: "Composer is installed", OK: composerErr == nil, Detail: "composer is not on PATH", Remedy: "install Composer (https://getcomposer.org/download)"},
	}}
}
