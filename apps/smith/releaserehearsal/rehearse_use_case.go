package releaserehearsal

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Rehearse asks every target's registry the four questions of a rehearsal:
// does the URL resolve and what credential does it expect, is the configured
// credential present, does the registry accept it, and in CI may the
// pipeline's identity publish. It publishes nothing and prints no secret. A
// target whose registry is unreachable, or whose credential is missing, is
// not asked further.
func Rehearse(targets []Target, probe *Probe, env func(string) string) Report {
	var report Report
	for _, t := range targets {
		rehearseOne(&report, t, probe, env)
	}
	return report
}

func rehearseOne(report *Report, t Target, probe *Probe, env func(string) string) {
	host := hostOf(t.URL)
	add := func(check, status, message, fix string) {
		report.Add(Result{Project: t.Project, Registry: host, Check: check, Status: status, Message: message, Fix: fix})
	}
	first, err := probe.Get(t.URL, "")
	if err != nil {
		add(CheckRegistry, Fail, fmt.Sprintf("%s is unreachable: %v", t.URL, err), "check the network and the URL; set it in smith.toml ([registries.<name>] and registry = \"<name>\") or in .npmrc")
		return
	}
	if first.Status >= 500 {
		add(CheckRegistry, Warn, fmt.Sprintf("%s answered %d, so it may be down", t.URL, first.Status), "check the registry's status page and run the rehearsal again")
	} else {
		add(CheckRegistry, Pass, fmt.Sprintf("%s resolves and expects %s", t.URL, expectation(t, first)), "")
	}
	cred := CredentialOf(t, env)
	provider := CIProvider(env)
	if cred.Value == "" {
		missing(add, t, cred, provider, env)
		return
	}
	add(CheckCredential, Pass, "the credential is set in "+cred.Source, "")
	status, message := askAccess(t, probe, first.Challenge, cred.Value)
	if status == Fail {
		add(CheckAccess, Fail, message, rejectFix(t, cred))
		return
	}
	add(CheckAccess, status, message, "")
	if provider != "" {
		add(CheckIdentity, Pass, identity(provider, cred), "")
	}
}

// expectation says what credential the first, unauthenticated answer asked for.
func expectation(t Target, a Answer) string {
	if a.Challenge.Scheme != "" {
		return a.Challenge.Expects()
	}
	switch t.Ecosystem {
	case PyPI:
		return "Basic with the user __token__ and an API token as the password"
	case NuGet:
		return "an API key in the X-NuGet-ApiKey header"
	case Crates:
		return "an API token in the Authorization header"
	}
	return "a token to publish (reads are open)"
}

// missing reports an absent credential. Some absences are expected: the
// pipeline signs in itself (Azure outside Azure), or a GitHub job holds the
// OIDC grant that npm, PyPI and crates.io accept instead of a token.
func missing(add func(check, status, message, fix string), t Target, c Credential, provider string, env func(string) string) {
	switch {
	case t.Auth == protocol.AuthAzureArtifacts && provider != "azure":
		add(CheckCredential, Warn, "Azure Artifacts signs in with the pipeline's token, which only exists in Azure Pipelines, so the credential was not checked here", "run the rehearsal in the pipeline, or put the feed's token in .npmrc to check it here")
	case t.Auth == "" && OIDCAvailable(env):
		add(CheckCredential, Warn, c.Want+" is not set; this job can sign in with OIDC (trusted publishing), which cannot be checked without publishing", "register this repository and workflow as a trusted publisher on "+t.Ecosystem+", or set "+c.Want)
	default:
		add(CheckCredential, Fail, missingMessage(t, c), setFix(t, c, provider))
	}
}

func missingMessage(t Target, c Credential) string {
	if c.Want == "" {
		return "smith.toml names no token variable for this registry (auth = \"token\" needs token_env)"
	}
	return fmt.Sprintf("%s is not set, so %s cannot be published to %s", c.Want, strings.Join(t.Names, " and "), hostOf(t.URL))
}

func setFix(t Target, c Credential, provider string) string {
	switch {
	case c.Want == "":
		return "set token_env under [registries.<name>] in smith.toml"
	case provider == "azure" && c.Want == "SYSTEM_ACCESSTOKEN":
		return "map the pipeline token into the release step: env: SYSTEM_ACCESSTOKEN: $(System.AccessToken)"
	case provider == "github" && c.Want == "GITHUB_TOKEN":
		return "pass the workflow token to the release step (env: GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}) and grant permissions: packages: write"
	case provider == "github" && t.Auth == "":
		return "add the secret " + c.Want + " to the repository, pass it to the release step, or grant permissions: id-token: write for trusted publishing"
	case provider != "":
		return "add the secret " + c.Want + " to the pipeline and pass it to the release step (smith ci generate does)"
	}
	return "export " + c.Want + " with a token that can publish " + strings.Join(t.Names, " and ")
}

func rejectFix(t Target, c Credential) string {
	switch t.Auth {
	case protocol.AuthAzureArtifacts:
		return "give the pipeline's Build Service identity the Contributor role on the feed (Azure DevOps > Artifacts > Feed settings > Permissions)"
	case protocol.AuthGitHubPackages:
		return "grant the job permissions: packages: write, and publish from a repository that owns the package"
	}
	return "create a token that can publish " + strings.Join(t.Names, " and ") + " and set " + c.Source
}

// identity says who publishes in CI, without the secret.
func identity(provider string, c Credential) string {
	return fmt.Sprintf("%s publishes with %s", map[string]string{"github": "the GitHub Actions job", "azure": "the Azure Pipelines job"}[provider], c.Source)
}

// askAccess makes the read-only authenticated request that tells whether the
// registry accepts the credential.
func askAccess(t Target, probe *Probe, c Challenge, token string) (status, message string) {
	switch t.Ecosystem {
	case NPM:
		return accessNPM(t, probe, c, token)
	case PyPI:
		a, err := probe.EmptyUpload(t.URL, "__token__", token)
		return verdict(a, err, "the index accepted the credential (an upload with no file, so nothing is published)")
	case NuGet:
		if c.Scheme == "" {
			return Pass, "the feed cannot check a push key without pushing; the key is set and sent only by the real push"
		}
		a, err := probe.Get(t.URL, Authorization(c, token))
		return verdict(a, err, "the feed accepted the credential")
	}
	a, err := probe.Get(strings.TrimSuffix(t.URL, "/")+"/api/v1/me", token)
	return verdict(a, err, "crates.io accepted the token")
}

func accessNPM(t Target, probe *Probe, c Challenge, token string) (string, string) {
	base := strings.TrimSuffix(t.URL, "/")
	auth := Authorization(c, token)
	a, err := probe.Get(base+"/-/whoami", auth)
	if err == nil && a.Status == http.StatusOK {
		if u := userOf(a.Body); u != "" {
			return Pass, "the registry accepted the token for " + u
		}
		return Pass, "the registry accepted the token"
	}
	if err == nil && a.Status == http.StatusNotFound {
		// Azure Artifacts and GitHub Packages have no /-/whoami: read a package.
		a, err = probe.Get(base+"/"+strings.ReplaceAll(url.PathEscape(t.Names[0]), "/", "%2f"), auth)
		if err == nil && a.Status == http.StatusNotFound {
			return Pass, "the registry accepted the token (the package is new)"
		}
	}
	return verdict(a, err, "the registry accepted the token")
}

func verdict(a Answer, err error, ok string) (string, string) {
	switch {
	case err != nil:
		return Fail, "the registry could not be asked: " + err.Error()
	case a.Status == http.StatusUnauthorized || a.Status == http.StatusForbidden:
		return Fail, fmt.Sprintf("the registry refused the credential (%d)", a.Status)
	case a.Status >= 500:
		return Warn, fmt.Sprintf("the registry answered %d, so the credential was not checked", a.Status)
	}
	return Pass, ok
}

func hostOf(raw string) string {
	if u, err := url.Parse(raw); err == nil && u.Host != "" {
		return u.Host
	}
	return raw
}
