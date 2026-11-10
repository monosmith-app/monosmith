package cachesetup

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/cachecredentials"
)

// ProviderGuide is what one CI provider needs for the remote cache: the
// secrets to create and the steps, in the provider's own words.
type ProviderGuide struct {
	// Provider is the name in `[ci] providers`.
	Provider string
	// Title is the provider's display name.
	Title string
	// Secrets name each secret to create, with what it is.
	Secrets []string
	// Steps are the steps for the web pages, in order.
	Steps []string
	// KeyCommand is a command that stores SMITH_CACHE_KEY, with the value on
	// standard input; nil when the provider has none smith can run.
	KeyCommand []string
	// ByHand is a command smith does not run, for the person to run.
	ByHand string
	// Notes are facts about this provider's pull requests and forks.
	Notes []string
}

// secretLines names each secret the remote needs, with what it is.
func secretLines(remote string) []string {
	var lines []string
	for _, c := range cachecredentials.For(remote) {
		lines = append(lines, fmt.Sprintf("%s: %s", c.Env, c.Description))
	}
	return lines
}

// ProviderGuides gives the instructions for each provider in `[ci] providers`.
//
// Parameters: providers are the names from smith.toml; remote is `[cache]
// remote`. An unknown name is skipped.
func ProviderGuides(providers []string, remote string) []ProviderGuide {
	var guides []ProviderGuide
	for _, p := range providers {
		g := ProviderGuide{Provider: p}
		secrets := secretLines(remote)
		switch p {
		case "github":
			g.Title = "GitHub Actions"
			g.Steps = []string{"Open the repository on GitHub: Settings, Secrets and variables, Actions, New repository secret.", "Create one secret for each name below, with the value as the secret."}
			g.KeyCommand = []string{"gh", "secret", "set", cachecredentials.Key}
			g.Notes = []string{"A pull request from a fork gets no secrets, so the remote is off there. Other pull requests read the cache and never write it."}
		case "azure":
			g.Title = "Azure Pipelines"
			g.Steps = []string{"Open the pipeline and choose Edit, then Variables, then New variable.", "Create one variable for each name below, and tick \"Keep this value secret\".", "smith ci generate already maps each secret into the steps with env:, because Azure exposes a secret only that way."}
			g.ByHand = "az pipelines variable create --name SMITH_CACHE_KEY --secret true --pipeline-name <your pipeline>   (it asks for the value; check `az pipelines variable create --help`)"
			g.Notes = []string{"In a build of a fork Azure leaves the variable's name in place of its value. smith reads that as unset, so the remote is off. Pull requests read the cache and never write it."}
		case "gitlab":
			g.Title = "GitLab CI"
			g.Steps = []string{"Open the project: Settings, CI/CD, Variables, Add variable.", "Create one variable for each name below. Turn on Masked. Leave Protected off, or merge request pipelines will not see it."}
			g.KeyCommand = []string{"glab", "variable", "set", cachecredentials.Key, "--masked"}
			g.Notes = []string{"Merge request pipelines read the cache and never write it."}
		case "circleci":
			g.Title = "CircleCI"
			g.Steps = []string{"Open the project: Project Settings, Environment Variables, Add Environment Variable (or put them in a context the project uses).", "Create one variable for each name below."}
			g.Notes = []string{"CircleCI withholds secrets from builds of forks by default. Pull requests read the cache and never write it."}
		case "bitbucket":
			g.Title = "Bitbucket Pipelines"
			g.Steps = []string{"Open the repository: Repository settings, Pipelines, Repository variables.", "Create one variable for each name below and tick Secured."}
			g.Notes = []string{"Pull requests read the cache and never write it."}
		default:
			continue
		}
		g.Secrets = secrets
		guides = append(guides, g)
	}
	return guides
}
