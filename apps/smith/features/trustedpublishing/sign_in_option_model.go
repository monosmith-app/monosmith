package trustedpublishing

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

// Place is what smith knows about where the workflow runs: the values a
// registry's trust policy names.
type Place struct {
	// Owner and Repository are the GitHub owner and repository name; empty
	// when the release remote is not a GitHub repository.
	Owner, Repository string
	// Workflow is the workflow file name, ci.yml.
	Workflow string
	// Branches are the release branch and the release channels' branches.
	Branches []string
}

// owner is the owner for the text, or a placeholder when it is not known.
func (p Place) owner() string { return orPlaceholder(p.Owner, "<owner>") }

// repository is the repository name for the text, or a placeholder.
func (p Place) repository() string { return orPlaceholder(p.Repository, "<repository>") }

func orPlaceholder(value, placeholder string) string {
	if value == "" {
		return placeholder
	}
	return value
}

// branches names the branches a federated credential is needed for.
func branches(names []string) string {
	if len(names) == 1 {
		return "branch " + names[0]
	}
	return "one credential for each of the branches " + strings.Join(names, ", ")
}

// Option is one registry sign-in that can replace a long-lived secret.
type Option struct {
	// Registry is the registry's name for the text ("NuGet").
	Registry string
	// Setting is the [release] key that records the choice.
	Setting string
	// Secret is the long-lived secret trusted publishing replaces.
	Secret string
	// Method names the keyless sign-in ("trusted publishing").
	Method string
	// steps lists what the person sets up once for the keyless sign-in.
	steps func(Place) []string
	// decided reads the setting: nil when not decided.
	decided func(workspace.ReleaseSection) *bool
	// record stores an answer in the section.
	record func(*workspace.ReleaseSection, bool)
}

// Path is the setting's path for smith config set ("release.nuget_trusted_publishing").
func (o Option) Path() string { return "release." + o.Setting }

// Decided reports whether smith.toml records an answer for the option.
//
// Parameters: r is the [release] section.
// Returns: true when the setting is present, true or false.
func (o Option) Decided(r workspace.ReleaseSection) bool { return o.decided(r) != nil }

// Steps is what the person sets up once for the keyless sign-in, with the
// values smith knows.
//
// Parameters: p is where the workflow runs.
// Returns: one sentence per step.
func (o Option) Steps(p Place) []string { return o.steps(p) }

// Options are the registries smith offers a keyless sign-in for, in the order
// it asks.
var Options = []Option{
	{
		Registry: "NuGet", Setting: "nuget_trusted_publishing", Secret: "NUGET_API_KEY", Method: "trusted publishing",
		steps: func(p Place) []string {
			return []string{
				fmt.Sprintf("On nuget.org, open your profile, then Trusted Publishing, and add a GitHub Actions policy: owner %s, repository %s, workflow file %s, no environment.", p.owner(), p.repository(), p.Workflow),
				"Add the repository secret NUGET_USER: your nuget.org profile name, not your email.",
			}
		},
		decided: func(r workspace.ReleaseSection) *bool { return r.NuGetTrustedPublishing },
		record:  func(r *workspace.ReleaseSection, v bool) { r.NuGetTrustedPublishing = &v },
	},
	{
		Registry: "RubyGems", Setting: "rubygems_trusted_publishing", Secret: "GEM_HOST_API_KEY", Method: "trusted publishing",
		steps: func(p Place) []string {
			return []string{
				fmt.Sprintf("On each gem's page on rubygems.org, open Trusted publishers and add GitHub Actions: owner %s, repository %s, workflow file %s, no environment. A gem that is not published yet takes a pending trusted publisher in your account settings.", p.owner(), p.repository(), p.Workflow),
			}
		},
		decided: func(r workspace.ReleaseSection) *bool { return r.RubyGemsTrustedPublishing },
		record:  func(r *workspace.ReleaseSection, v bool) { r.RubyGemsTrustedPublishing = &v },
	},
	{
		Registry: "the VS Code Marketplace", Setting: "vsce_entra", Secret: "VSCE_PAT", Method: "Microsoft Entra ID",
		steps: func(p Place) []string {
			return []string{
				fmt.Sprintf("In Microsoft Entra ID, register an app and add a federated credential for GitHub Actions: organisation %s, repository %s, entity Branch, %s.", p.owner(), p.repository(), branches(p.Branches)),
				"On marketplace.visualstudio.com/manage, add the app to the publisher's members with the Contributor role.",
				"Add the repository secrets AZURE_CLIENT_ID and AZURE_TENANT_ID: they name the app and are not credentials.",
			}
		},
		decided: func(r workspace.ReleaseSection) *bool { return r.VSCEEntra },
		record:  func(r *workspace.ReleaseSection, v bool) { r.VSCEEntra = &v },
	},
}
