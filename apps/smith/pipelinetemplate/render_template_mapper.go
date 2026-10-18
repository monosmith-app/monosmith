package pipelinetemplate

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
)

// App is the project a deploy template builds.
type App struct {
	Name string
	Root string
	// Output is its build task's output directory, project-relative.
	Output string
}

// Path is where a template's file goes for a provider.
func Path(name, provider string) string {
	if provider == "azure" {
		return "azure-pipelines/" + name + ".yml"
	}
	return ".github/workflows/" + name + ".yml"
}

func written(name, extra string) string {
	return fmt.Sprintf("# Written by `smith pipeline %s%s`. Yours from here: smith upgrade does not rewrite it.\n", name, extra)
}

// Render writes a template for a provider; ok is false when the template has
// no form for it.
func Render(name, provider string, m cipipeline.Model, app App) (string, bool) {
	project := ""
	if app.Name != "" {
		project = " --project " + app.Name
	}
	githubOn := fmt.Sprintf("on:\n  push:\n    branches: [%s]\n  workflow_dispatch:\n", m.ReleaseBranch)
	azureOn := fmt.Sprintf("trigger:\n  branches:\n    include:\n      - %s\npr: none\n\npool:\n  vmImage: ubuntu-latest\n", m.ReleaseBranch)
	azureNote := "# Register it as a pipeline of its own (Pipelines → New pipeline → Existing YAML file).\n"
	switch name + "/" + provider {
	case "e2e/github":
		return "name: e2e\n\n" + written(name, "") +
			"# Runs every project's e2e task, on pull requests too. It is not part of smith ci\n" +
			"# verify, because it needs what verify does not install: the playwright plugin's\n" +
			"# e2e-browsers task installs the browsers, cached here between runs. The suites run\n" +
			"# one at a time, since they share ports and, under xvfb-run, a display.\n\n" + githubOn + "  pull_request:\n" +
			"\njobs:\n  e2e:\n    runs-on: ubuntu-latest\n    steps:\n" + cipipeline.GitHubSetup(m) +
			"      - uses: actions/cache@v4\n        with:\n          path: ~/.cache/ms-playwright\n" +
			// The lockfile fixes the Playwright version; the manifests name the browsers.
			"          key: playwright-${{ runner.os }}-${{ hashFiles('package-lock.json', '**/package.json', '!**/node_modules/**') }}\n" +
			"      - name: End-to-end tests\n        run: smith run e2e --parallel 1\n", true
	case "e2e/azure":
		return written(name, "") + azureNote + "# Runs every project's e2e task, which smith ci verify leaves out, one at a time:\n" +
			"# the suites share ports and, under xvfb-run, a display. The pr trigger works for a\n" +
			"# GitHub repository; on Azure Repos, add this pipeline as a build validation policy.\n\n" +
			strings.Replace(azureOn, "pr: none\n", "pr:\n  branches:\n    include:\n      - "+m.ReleaseBranch+"\n", 1) +
			"\nvariables:\n  PLAYWRIGHT_BROWSERS_PATH: $(Pipeline.Workspace)/ms-playwright\n" +
			"\nsteps:\n" + cipipeline.AzureSetup(m) +
			"  - task: Cache@2\n    displayName: Cache the Playwright browsers\n    inputs:\n" +
			"      key: 'playwright | \"$(Agent.OS)\" | package-lock.json | **/package.json, !**/node_modules/**'\n" +
			"      path: $(PLAYWRIGHT_BROWSERS_PATH)\n" +
			"  - bash: smith run e2e --parallel 1\n    displayName: End-to-end tests\n", true
	case "package-zip/github":
		return "name: package\n\n" + written(name, "") + "# Packs the projects into dist/drop and keeps it as the drop artifact.\n\n" + githubOn +
			"\njobs:\n  package:\n    runs-on: ubuntu-latest\n    steps:\n" + cipipeline.GitHubSetup(m) +
			"      - name: Pack\n        run: smith ci pack\n" +
			"      - uses: actions/upload-artifact@v7\n        with:\n          name: drop\n          path: " + cipipeline.DropDir + "\n          if-no-files-found: error\n", true
	case "package-zip/azure":
		return written(name, "") + azureNote + "# The drop artifact is what a classic release pipeline takes its packages from.\n\n" + azureOn +
			"\nsteps:\n" + cipipeline.AzureSetup(m) + "  - bash: smith ci pack\n    displayName: Pack\n" +
			"  - task: PublishBuildArtifacts@1\n    displayName: Publish dist/drop as the drop artifact\n    inputs:\n      PathtoPublish: " + cipipeline.DropDir + "\n      ArtifactName: drop\n", true
	case "deploy-pages/github":
		return "name: deploy " + app.Name + " to GitHub Pages\n\n" + written(name, project) +
			"# Needs Settings → Pages → Source set to \"GitHub Actions\". A project site is\n" +
			"# served under /<repository>/: build the app with that base path if it needs one.\n\n" + githubOn +
			"\npermissions:\n  contents: read\n  pages: write\n  id-token: write\n\nconcurrency:\n  group: pages\n  cancel-in-progress: false\n" +
			"\njobs:\n  deploy:\n    runs-on: ubuntu-latest\n    environment:\n      name: github-pages\n      url: ${{ steps.deployment.outputs.page_url }}\n    steps:\n" + cipipeline.GitHubSetup(m) +
			"      - uses: actions/configure-pages@v5\n" +
			"      - name: Build " + app.Name + "\n        run: smith run build -p " + app.Name + "\n" +
			"      - uses: actions/upload-pages-artifact@v3\n        with:\n          path: " + app.Root + "/" + app.Output + "\n" +
			"      - uses: actions/deploy-pages@v4\n        id: deployment\n", true
	case "deploy-azure-function/github":
		return "name: deploy " + app.Name + " to Azure\n\n" + written(name, project) +
			"# Signs in with OIDC, no stored secret: the repository variables AZURE_CLIENT_ID,\n" +
			"# AZURE_TENANT_ID and AZURE_SUBSCRIPTION_ID name an Entra app whose federated\n" +
			"# credential trusts this repository's release branch, and AZURE_FUNCTIONAPP_NAME\n" +
			"# names the Function App.\n\n" + githubOn +
			"\npermissions:\n  contents: read\n  id-token: write\n" +
			"\njobs:\n  deploy:\n    runs-on: ubuntu-latest\n    steps:\n" + cipipeline.GitHubSetup(m) +
			"      - name: Build " + app.Name + "\n        run: smith run build -p " + app.Name + "\n" +
			"      - uses: azure/login@v2\n        with:\n          client-id: ${{ vars.AZURE_CLIENT_ID }}\n          tenant-id: ${{ vars.AZURE_TENANT_ID }}\n          subscription-id: ${{ vars.AZURE_SUBSCRIPTION_ID }}\n" +
			"      - uses: azure/functions-action@v1\n        with:\n          app-name: ${{ vars.AZURE_FUNCTIONAPP_NAME }}\n          package: " + app.Root + "\n", true
	case "deploy-azure-function/azure":
		return written(name, project) + azureNote +
			"# Needs a service connection (AZURE_SERVICE_CONNECTION) and the Function App's\n" +
			"# name (AZURE_FUNCTIONAPP_NAME) as pipeline variables.\n\n" + azureOn +
			"\nsteps:\n" + cipipeline.AzureSetup(m) + "  - bash: smith run build -p " + app.Name + "\n    displayName: Build " + app.Name + "\n" +
			"  - task: AzureFunctionApp@2\n    displayName: Deploy " + app.Name + "\n    inputs:\n      connectedServiceNameARM: $(AZURE_SERVICE_CONNECTION)\n      appType: functionAppLinux\n      appName: $(AZURE_FUNCTIONAPP_NAME)\n      package: $(System.DefaultWorkingDirectory)/" + app.Root + "\n", true
	}
	return "", false
}
