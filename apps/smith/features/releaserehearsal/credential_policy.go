package releaserehearsal

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// Credential is the secret a publish to a target signs in with. Value is
// never printed or encoded; Source names where it came from.
type Credential struct {
	Value  string
	Source string
	// Want is the variable or setting to give when Value is empty.
	Want string
}

// CIProvider names the pipeline that runs this process: github, azure or "".
func CIProvider(env func(string) string) string {
	switch {
	case env("GITHUB_ACTIONS") == "true":
		return "github"
	case env("TF_BUILD") != "":
		return "azure"
	}
	return ""
}

// OIDCAvailable is true in a GitHub Actions job granted `id-token: write`,
// which npm, PyPI and crates.io accept for trusted publishing.
func OIDCAvailable(env func(string) string) bool {
	return CIProvider(env) == "github" && env("ACTIONS_ID_TOKEN_REQUEST_URL") != ""
}

// CredentialOf finds the credential a target's publish signs in with, using
// the variables the language plugins read: the registry's own sign-in from
// smith.toml first, else the ecosystem's usual token variables.
func CredentialOf(t Target, env func(string) string) Credential {
	switch t.Auth {
	case protocol.AuthAzureArtifacts:
		if v := env("SYSTEM_ACCESSTOKEN"); v != "" {
			return Credential{Value: v, Source: "SYSTEM_ACCESSTOKEN"}
		}
		if t.NpmrcToken != "" {
			return Credential{Value: t.NpmrcToken, Source: ".npmrc"}
		}
		return Credential{Want: "SYSTEM_ACCESSTOKEN"}
	case protocol.AuthGitHubPackages:
		return pick(env, "GITHUB_TOKEN")
	case protocol.AuthToken:
		return pick(env, t.TokenEnv)
	}
	switch t.Ecosystem {
	case NPM:
		if c := pick(env, "NODE_AUTH_TOKEN", "NPM_TOKEN"); c.Value != "" {
			return c
		}
		if t.NpmrcToken != "" {
			return Credential{Value: t.NpmrcToken, Source: ".npmrc"}
		}
		return Credential{Want: "NODE_AUTH_TOKEN"}
	case PyPI:
		return pick(env, "PYPI_TOKEN", "TWINE_PASSWORD")
	case NuGet:
		return pick(env, "NUGET_API_KEY")
	}
	return pick(env, "CARGO_REGISTRY_TOKEN")
}

func pick(env func(string) string, names ...string) Credential {
	for _, n := range names {
		if n != "" && env(n) != "" {
			return Credential{Value: env(n), Source: n}
		}
	}
	return Credential{Want: names[0]}
}
