package images

import (
	"net"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Login is how release authenticates to a registry.
type Login int

const (
	// LoginPassword logs in with the registry's username and password
	// variables: CONTAINER_USERNAME and CONTAINER_PASSWORD, or the ones the
	// registry's entry names.
	LoginPassword Login = iota
	// LoginGitHub logs in to ghcr.io with the job's GITHUB_TOKEN.
	LoginGitHub
	// LoginNone pushes without logging in: a registry on this machine.
	LoginNone
	// LoginDockerHub logs in to Docker Hub with DOCKERHUB_USERNAME and
	// DOCKERHUB_TOKEN.
	LoginDockerHub
)

// Host is the registry's host (ghcr.io, localhost:5000).
func Host(registry string) string {
	host, _, _ := strings.Cut(registry, "/")
	return host
}

// Loopback reports whether a registry is on this machine.
func Loopback(registry string) bool {
	name := Host(registry)
	if h, _, err := net.SplitHostPort(name); err == nil {
		name = h
	}
	return name == "localhost" || net.ParseIP(strings.Trim(name, "[]")).IsLoopback()
}

// LoginFor decides how release logs in to a registry that names no
// credentials of its own.
func LoginFor(registry string) Login {
	switch Host(registry) {
	case "ghcr.io":
		return LoginGitHub
	case "docker.io", "index.docker.io", "registry-1.docker.io":
		return LoginDockerHub
	}
	if Loopback(registry) {
		return LoginNone
	}
	return LoginPassword
}

// Registry is one registry an image is pushed to, with how release logs in.
type Registry struct {
	// URL is the registry and namespace (ghcr.io/acme), no trailing slash.
	URL   string
	Login Login
	// UserEnv and PasswordEnv name the variables that hold the credentials.
	// Empty for LoginNone; for LoginGitHub UserEnv is optional (the actor).
	UserEnv, PasswordEnv string
	// Explicit is true when smith.toml names the variables.
	Explicit bool
}

// Host is the registry's host.
func (r Registry) Host() string { return Host(r.URL) }

// Credentials are the secrets pushing to the targets needs, and the GitHub
// token permissions. A registry set with CONTAINER_REGISTRY keeps that
// variable's old, optional credentials. Several registries that share a
// variable (two docker.io namespaces) list it once.
func Credentials(t Target) ([]protocol.Credential, []string) {
	if t.Legacy {
		return []protocol.Credential{
			{Env: LegacyRegistryEnv, Description: "Registry the apps' images are pushed to; deprecated: set [conventions.docker] registry in smith.toml"},
			{Env: "CONTAINER_USERNAME", Description: "Registry user (omit to use the runner's docker login)"},
			{Env: "CONTAINER_PASSWORD", Description: "Registry password or token"},
		}, nil
	}
	if !t.Push {
		return nil, nil
	}
	var creds []protocol.Credential
	var perms []string
	seen := map[string]bool{}
	add := func(env, description string) {
		if !seen[env] {
			seen[env] = true
			creds = append(creds, protocol.Credential{Env: env, Description: description, Required: true})
		}
	}
	for _, r := range t.Registries {
		switch r.Login {
		case LoginNone:
		case LoginGitHub:
			add("GITHUB_TOKEN", "pushes the images to "+r.URL+" ([conventions.docker] registries): the job's token with packages: write; on Azure Pipelines, a token with write:packages")
			if !seen["packages: write"] {
				seen["packages: write"] = true
				perms = append(perms, "packages: write")
			}
		default:
			add(r.UserEnv, "user that pushes the images to "+r.URL+" ([conventions.docker] registries)")
			add(r.PasswordEnv, "password or token that pushes the images to "+r.URL+" ([conventions.docker] registries)")
		}
	}
	return creds, perms
}
