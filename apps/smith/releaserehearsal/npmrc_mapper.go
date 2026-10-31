package releaserehearsal

import (
	"net/url"
	"os"
	"regexp"
	"strings"
)

var envRef = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// NpmrcRegistry reads the registry npm publishes name to from .npmrc text:
// the `@scope:registry` line when name is scoped, else `registry`. A later
// line wins. The empty string means the text names none.
func NpmrcRegistry(npmrc, name string) string {
	scope := ""
	if strings.HasPrefix(name, "@") {
		scope, _, _ = strings.Cut(name, "/")
	}
	var scoped, global string
	for _, line := range strings.Split(npmrc, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.Trim(strings.TrimSpace(value), `"'`)
		switch {
		case scope != "" && key == scope+":registry":
			scoped = value
		case key == "registry":
			global = value
		}
	}
	if scoped != "" {
		return scoped
	}
	return global
}

// NpmrcToken finds the `//host/path/:_authToken=` (or `:_password=`) line that
// covers registry and expands a ${VAR} in it with env. It returns the token and
// the variable the line names ("" for a literal value). The longest matching
// prefix wins.
func NpmrcToken(npmrc, registry string, env func(string) string) (token, variable string) {
	u, err := url.Parse(registry)
	if err != nil || u.Host == "" {
		return "", ""
	}
	want := "//" + u.Host + strings.TrimSuffix(u.Path, "/") + "/"
	best := -1
	for _, line := range strings.Split(npmrc, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		key = strings.TrimSpace(key)
		if !ok || !strings.HasSuffix(key, ":_authToken") && !strings.HasSuffix(key, ":_password") {
			continue
		}
		prefix := strings.TrimSuffix(strings.TrimSuffix(key, ":_authToken"), ":_password")
		if !strings.HasPrefix(want, prefix) || len(prefix) <= best {
			continue
		}
		best = len(prefix)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		variable = ""
		if m := envRef.FindStringSubmatch(value); m != nil {
			variable = m[1]
		}
		token = envRef.ReplaceAllStringFunc(value, func(ref string) string { return env(envRef.FindStringSubmatch(ref)[1]) })
	}
	return token, variable
}

// readNpmrc joins the .npmrc files npm reads, lowest priority first, so the
// last line for a key is the one npm uses.
func readNpmrc(paths ...string) string {
	var out []string
	for _, p := range paths {
		if raw, err := os.ReadFile(p); err == nil {
			out = append(out, string(raw))
		}
	}
	return strings.Join(out, "\n")
}
