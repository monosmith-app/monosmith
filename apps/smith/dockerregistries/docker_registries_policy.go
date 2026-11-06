package workspace

import (
	"fmt"
	"strings"
)

// ValidRegistryURL reports why a "host/namespace" registry is malformed, or
// "" when it is fine. The host is the first segment and must look like one (a
// dot, a port, or localhost): Docker reads a name without one as a Docker Hub
// namespace, so docker.io/acme must say docker.io.
func ValidRegistryURL(url string) string {
	switch {
	case url == "" || strings.TrimSpace(url) != url || strings.ContainsAny(url, " \t\\?#@"):
		return "must be host/namespace, with no spaces"
	case strings.Contains(url, "://"):
		return "must not start with a scheme: write ghcr.io/acme, not https://ghcr.io/acme"
	case url != strings.ToLower(url):
		return "must be lowercase"
	case strings.Contains(url, ".."):
		return "must not contain .."
	}
	host, _, _ := strings.Cut(url, "/")
	if !strings.ContainsAny(host, ".:") && host != "localhost" {
		return fmt.Sprintf("must start with the registry's host: %q reads as a Docker Hub namespace (write docker.io/%s for Docker Hub)", host, url)
	}
	return ""
}

// checkDockerRegistries validates one table's registry settings
// ([conventions.docker] or [projects.<name>.docker]): registry and registries
// are alternatives, the list is well formed, and description_from stays in the
// workspace. It returns the normalized list, nil when none is set.
func checkDockerRegistries(label, registry string, registries []any, descriptionFrom string) ([]any, error) {
	if registry != "" && registries != nil {
		return nil, fmt.Errorf("%s sets both registry and registries: use registries for several, registry for one", label)
	}
	if descriptionFrom != "" && !ValidContext(descriptionFrom) {
		return nil, fmt.Errorf("%s description_from %q must be a workspace-relative file with forward slashes (README.md, say)", label, descriptionFrom)
	}
	if registries == nil {
		return nil, nil
	}
	return DockerRegistries(label+" registries", registries)
}

// DockerRegistries checks the docker convention's registries list and returns
// it in the form the plugin reads: a "host/namespace" string, or a table with
// url, username_env and password_env. label names the setting in errors.
func DockerRegistries(label string, raw []any) ([]any, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("%s is empty: list the registries, or leave it out and use registry for one", label)
	}
	seen := map[string]bool{}
	out := make([]any, 0, len(raw))
	for i, item := range raw {
		var url string
		switch v := item.(type) {
		case string:
			url = strings.TrimSuffix(v, "/")
			if why := ValidRegistryURL(url); why != "" {
				return nil, fmt.Errorf("%s[%d] %q %s", label, i, v, why)
			}
			out = append(out, url)
		case map[string]any:
			for key := range v {
				if key != "url" && key != "username_env" && key != "password_env" {
					return nil, fmt.Errorf("%s[%d] has an unknown key %q (use url, username_env and password_env)", label, i, key)
				}
			}
			url, _ = v["url"].(string)
			url = strings.TrimSuffix(url, "/")
			if why := ValidRegistryURL(url); why != "" {
				return nil, fmt.Errorf("%s[%d] url %q %s", label, i, url, why)
			}
			user, _ := v["username_env"].(string)
			pass, _ := v["password_env"].(string)
			if (user == "") != (pass == "") {
				return nil, fmt.Errorf("%s[%d] (%s) names one of username_env and password_env: name both, or neither for the host's default", label, i, url)
			}
			for _, name := range []string{user, pass} {
				if name != "" && !envName.MatchString(name) {
					return nil, fmt.Errorf("%s[%d] (%s): %q is not an environment variable name", label, i, url, name)
				}
			}
			entry := map[string]any{"url": url}
			if user != "" {
				entry["username_env"], entry["password_env"] = user, pass
			}
			out = append(out, entry)
		default:
			return nil, fmt.Errorf("%s[%d] must be a \"host/namespace\" string or a table with url, username_env and password_env", label, i)
		}
		if seen[url] {
			return nil, fmt.Errorf("%s lists %s twice", label, url)
		}
		seen[url] = true
	}
	return out, nil
}
