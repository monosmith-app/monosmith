package convertplan

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// planCISettings says what happens to the CI definitions the workspace has:
// smith writes its own, so what the old ones ran on and read is carried (an
// Azure agent pool and variable groups) and the rest is reported rather than
// dropped without a word.
func planCISettings(plan *Plan, in Input, has map[string]bool) {
	if contains(plan.Config.CI.Providers, "azure") && has["azure-pipelines.yml"] {
		pool, groups := AzurePipelineSettings(in.Read("azure-pipelines.yml"))
		plan.Config.CI.AzurePool, plan.Config.CI.AzureVariableGroups = pool, groups
		if pool != "" {
			plan.Add(Mapped, "", "azure-pipelines.yml", "runs on the agent pool "+pool+" ([ci] azure_pool)")
		}
		if len(groups) > 0 {
			plan.Add(Mapped, "", "azure-pipelines.yml", "links the variable group(s) "+strings.Join(groups, ", ")+" ([ci] azure_variable_groups)")
		}
		if name, url, ok := NpmrcFeed(in.Read(".npmrc")); ok && plan.Config.Registries == nil {
			plan.Config.Registries = map[string]workspace.RegistryConfig{name: {Auth: "azure-artifacts", NPM: url}}
			plan.Config.Release.Registry = name
			plan.Add(Mapped, "", ".npmrc", "routes a scope to the Azure Artifacts feed "+url+": [registries."+name+"] (auth = azure-artifacts, signed in by the pipeline) and [release] registry = "+name+", where the packages publish; installs stay as they were")
		}
		plan.Decide("", "azure-pipelines.yml", "replaced by the pipeline smith generates: steps of your own are not carried (see git diff), put them between the smith:slot markers; a private npm feed other than Azure Artifacts needs a [registries] entry")
	}
	if contains(plan.Config.CI.Providers, "github") && (has[".github/workflows/ci.yml"] || has[".github/workflows/ci.yaml"]) {
		plan.Decide("", ".github/workflows/ci.yml", "replaced by the workflow smith generates: steps of your own are not carried (see git diff), put them between the smith:slot markers")
	}
}
