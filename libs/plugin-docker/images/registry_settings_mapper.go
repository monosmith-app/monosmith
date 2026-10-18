package images

import (
	"fmt"
	"strings"
)

// RegistriesOf reads the registries the docker settings name: registry (one)
// or registries (a list of "host/namespace" strings or tables with url,
// username_env and password_env). With neither, CONTAINER_REGISTRY from env
// still names one (legacy). Each registry gets the credentials its host
// implies, unless its entry names them.
//
// Two registries on hosts with no default (neither ghcr.io, docker.io nor
// this machine) would both read CONTAINER_USERNAME and CONTAINER_PASSWORD, so
// each must name its own variables.
func RegistriesOf(s map[string]any, env func(string) string) (regs []Registry, legacy bool, err error) {
	single, _ := s["registry"].(string)
	list, hasList := s["registries"]
	if single != "" && hasList {
		return nil, false, fmt.Errorf("[conventions.docker] sets both registry and registries: use registries for several, registry for one")
	}
	var entries []registryEntry
	switch {
	case hasList:
		entries, err = entriesOf(list)
		if err != nil {
			return nil, false, err
		}
	case single != "":
		entries = []registryEntry{{url: single}}
	default:
		if old := env(LegacyRegistryEnv); old != "" {
			entries, legacy = []registryEntry{{url: old}}, true
		}
	}
	var shared []string
	seen := map[string]bool{}
	for _, e := range entries {
		r := resolve(e)
		if seen[r.URL] {
			return nil, false, fmt.Errorf("registry %s is listed twice", r.URL)
		}
		seen[r.URL] = true
		if !r.Explicit && r.Login == LoginPassword {
			shared = append(shared, r.URL)
		}
		regs = append(regs, r)
	}
	if len(shared) > 1 {
		return nil, false, fmt.Errorf("registries %s would all read CONTAINER_USERNAME and CONTAINER_PASSWORD: give each its own, as { url = %q, username_env = \"…\", password_env = \"…\" } in [conventions.docker] registries",
			strings.Join(shared, " and "), shared[0])
	}
	return regs, legacy, nil
}

// registryEntry is one element of the registries list, before its host
// decides the credentials.
type registryEntry struct {
	url, userEnv, passwordEnv string
}

// entriesOf reads the list: strings, or tables, as a TOML list decoded to
// JSON arrives.
func entriesOf(raw any) ([]registryEntry, error) {
	var items []any
	switch v := raw.(type) {
	case []any:
		items = v
	case []string:
		for _, s := range v {
			items = append(items, s)
		}
	default:
		return nil, fmt.Errorf("[conventions.docker] registries must be a list")
	}
	var out []registryEntry
	for i, item := range items {
		switch v := item.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return nil, fmt.Errorf("[conventions.docker] registries[%d] is empty", i)
			}
			out = append(out, registryEntry{url: strings.TrimSuffix(strings.TrimSpace(v), "/")})
		case map[string]any:
			url, _ := v["url"].(string)
			user, _ := v["username_env"].(string)
			pass, _ := v["password_env"].(string)
			if url == "" {
				return nil, fmt.Errorf("[conventions.docker] registries[%d] needs a url", i)
			}
			if (user == "") != (pass == "") {
				return nil, fmt.Errorf("[conventions.docker] registries[%d] (%s) names one of username_env and password_env: name both", i, url)
			}
			out = append(out, registryEntry{url: strings.TrimSuffix(url, "/"), userEnv: user, passwordEnv: pass})
		default:
			return nil, fmt.Errorf("[conventions.docker] registries[%d] must be a \"host/namespace\" string or a table with url", i)
		}
	}
	return out, nil
}

// resolve gives a registry its credentials: the ones its entry names, or the
// host's defaults.
func resolve(e registryEntry) Registry {
	e.url = strings.TrimSuffix(strings.TrimSpace(e.url), "/")
	if e.userEnv != "" {
		return Registry{URL: e.url, Login: LoginPassword, UserEnv: e.userEnv, PasswordEnv: e.passwordEnv, Explicit: true}
	}
	r := Registry{URL: e.url, Login: LoginFor(e.url)}
	switch r.Login {
	case LoginGitHub:
		r.UserEnv, r.PasswordEnv = "GITHUB_ACTOR", "GITHUB_TOKEN"
	case LoginDockerHub:
		r.UserEnv, r.PasswordEnv = "DOCKERHUB_USERNAME", "DOCKERHUB_TOKEN"
	case LoginPassword:
		r.UserEnv, r.PasswordEnv = "CONTAINER_USERNAME", "CONTAINER_PASSWORD"
	}
	return r
}
