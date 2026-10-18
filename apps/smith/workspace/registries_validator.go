package workspace

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var (
	// envName is a portable environment variable name.
	envName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	// registryName works as a key in every client's configuration (a
	// nuget.config element name among them).
	registryName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]*$`)
)

// validateRegistries checks [registries] and every name that refers to one,
// and fills the PyPI upload URL's default.
func validateRegistries(cfg *Config) error {
	names := make([]string, 0, len(cfg.Registries))
	for name := range cfg.Registries {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		r := cfg.Registries[name]
		at := fmt.Sprintf("%s: [registries.%s]", FileName, name)
		if !registryName.MatchString(name) {
			return fmt.Errorf("%s: the name must start with a letter and hold only letters, digits, '.', '-' and '_'", at)
		}
		switch r.Auth {
		case "azure-artifacts", "github-packages":
			if r.TokenEnv != "" {
				return fmt.Errorf("%s token_env is only for auth = \"token\"", at)
			}
		case "token":
			if !envName.MatchString(r.TokenEnv) {
				return fmt.Errorf("%s auth = \"token\" needs token_env, the environment variable holding the token", at)
			}
		case "":
			return fmt.Errorf("%s needs auth: \"azure-artifacts\", \"github-packages\" or \"token\"", at)
		default:
			return fmt.Errorf("%s unknown auth %q: use \"azure-artifacts\", \"github-packages\" or \"token\"", at, r.Auth)
		}
		if r.NPM == "" && r.PyPI == "" && r.PyPIUpload == "" && r.NuGet == "" {
			return fmt.Errorf("%s gives no npm, pypi or nuget URL", at)
		}
		if r.Auth == "github-packages" && (r.PyPI != "" || r.PyPIUpload != "") {
			return fmt.Errorf("%s GitHub Packages has no Python registry", at)
		}
		for key, u := range map[string]string{"npm": r.NPM, "pypi": r.PyPI, "pypi_upload": r.PyPIUpload, "nuget": r.NuGet} {
			if u == "" {
				continue
			}
			if parsed, err := url.Parse(u); err != nil || parsed.Scheme != "https" || parsed.Host == "" {
				return fmt.Errorf("%s %s = %q must be an https URL", at, key, u)
			}
		}
		if r.PyPIUpload == "" && strings.HasSuffix(strings.TrimSuffix(r.PyPI, "/"), "/simple") {
			r.PyPIUpload = strings.TrimSuffix(strings.TrimSuffix(r.PyPI, "/"), "/simple") + "/upload/"
		}
		cfg.Registries[name] = r
	}
	known := func(where, name string) error {
		if name == "" {
			return nil
		}
		if _, ok := cfg.Registries[name]; !ok {
			return fmt.Errorf("%s: %s registry = %q names no [registries.%s]", FileName, where, name, name)
		}
		return nil
	}
	if err := known("[workspace]", cfg.Workspace.Registry); err != nil {
		return err
	}
	if err := known("[release]", cfg.Release.Registry); err != nil {
		return err
	}
	projects := make([]string, 0, len(cfg.Projects))
	for name := range cfg.Projects {
		projects = append(projects, name)
	}
	sort.Strings(projects)
	for _, name := range projects {
		if err := known("[projects."+name+"]", cfg.Projects[name].Registry); err != nil {
			return err
		}
	}
	return nil
}
