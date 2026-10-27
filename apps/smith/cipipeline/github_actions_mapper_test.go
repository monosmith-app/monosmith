package cipipeline

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/cachecredentials"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// runtime is a toolchain whose plugin gives both providers' steps.
func runtime(name, version, action, task, input string) protocol.Toolchain {
	return protocol.Toolchain{
		Name: name, Version: version,
		GitHub: &protocol.ActionStep{Uses: action, Inputs: []protocol.Input{{Name: name + "-version", Value: version}, {Name: "cache", Value: "false"}}},
		Azure:  &protocol.ActionStep{Uses: task, Inputs: []protocol.Input{{Name: input, Value: version + ".x"}}},
	}
}

// scriptModel is model() on a workspace with no Node, Python or .NET toolchain,
// which installs smith with the committed script.
func scriptModel() Model {
	m := model()
	m.Toolchains = slices.DeleteFunc(m.Toolchains, func(t protocol.Toolchain) bool {
		return t.Name == "node" || t.Name == "python" || t.Name == "dotnet"
	})
	return m
}

func model() Model {
	var c Contributions
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{runtime("python", "3.12", "actions/setup-python@v7", "UsePythonVersion@0", "versionSpec")}, Credentials: []protocol.Credential{{Env: "PYPI_TOKEN", Required: true}}})
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{runtime("node", "24", "actions/setup-node@v7", "UseNode@1", "version"), {Name: "python", Version: "9"}}, Credentials: []protocol.Credential{{Env: "NODE_AUTH_TOKEN"}}})
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{runtime("go", "1.26", "actions/setup-go@v7", "GoTool@0", "version")}})
	// A plugin for a language smith does not know: an install command only,
	// or nothing at all.
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{{Name: "zig", Version: "0.14", Install: "curl -fsSL https://example.com/zig.sh | sh\nzig version"}, {Name: "rust", Version: "1.90"}}})
	return NewModel(c, "dev", "main")
}

// An extension package's plugin declares VSCE_PAT; both providers pass it
// to the release step, where vsce publish reads it.
func TestTheReleaseStepGetsTheMarketplaceToken(t *testing.T) {
	var c Contributions
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{runtime("node", "24", "actions/setup-node@v7", "UseNode@1", "version")}, Credentials: []protocol.Credential{{Env: "VSCE_PAT", Required: true}}})
	m := NewModel(c, "dev", "main")
	gh := RenderGitHub(m)
	if release := gh[strings.Index(gh, "      - name: Release\n"):]; !strings.Contains(release, "          VSCE_PAT: ${{ secrets.VSCE_PAT }}\n") {
		t.Errorf("GitHub release step:\n%s", release)
	}
	az := RenderAzure(m)
	if release := az[strings.Index(az, "smith ci release"):]; !strings.Contains(release, "VSCE_PAT: $(VSCE_PAT)\n") {
		t.Errorf("Azure release step:\n%s", release)
	}
}

func TestRenderersCarryTheSameSteps(t *testing.T) {
	m := model()
	gh, az := RenderGitHub(m), RenderAzure(m)
	for _, want := range []string{
		"      - uses: actions/setup-node@v7\n        with:\n          node-version: '24'\n          cache: false\n",
		"python-version: '3.12'", "go-version: '1.26'",
		"      - name: Install zig\n        shell: bash\n        run: |\n          curl -fsSL https://example.com/zig.sh | sh\n          zig version\n",
		"      # smith: no plugin says how to install rust 1.90",
		"smith ci setup", "smith ci audit", "smith ci verify", "smith ci release", "smith release --dry-run", "NODE_AUTH_TOKEN: ${{ secrets.NODE_AUTH_TOKEN }}", "fetch-depth: 0", "run: |\n          npm install --global --registry https://registry.npmjs.org/ monosmith\n          smith --version\n", "github.ref_name == 'main'"} {
		if !strings.Contains(gh, want) {
			t.Errorf("GitHub workflow lacks %q", want)
		}
	}
	for _, want := range []string{
		"  - task: UseNode@1\n    inputs:\n      version: '24.x'\n",
		"versionSpec: '3.12.x'", "  - task: GoTool@0",
		"  - bash: |\n      curl -fsSL https://example.com/zig.sh | sh\n      zig version\n    displayName: Install zig\n",
		"  # smith: no plugin says how to install rust 1.90",
		"smith ci setup", "smith ci audit", "smith ci verify", "smith ci release", "smith release --dry-run", "PYPI_TOKEN: $(PYPI_TOKEN)", "fetchDepth: 0", "persistCredentials: true", "refs/heads/main"} {
		if !strings.Contains(az, want) {
			t.Errorf("Azure pipeline lacks %q", want)
		}
	}
	// The phase steps are the same in both providers: the lockstep rule.
	phases := func(s string) []string {
		var out []string
		for _, line := range strings.Split(s, "\n") {
			if i := strings.Index(line, "smith ci "); i >= 0 {
				out = append(out, strings.TrimSpace(line[i:]))
			}
		}
		return out
	}
	if strings.Join(phases(gh), ",") != strings.Join(phases(az), ",") {
		t.Fatalf("providers drifted:\n%v\n%v", phases(gh), phases(az))
	}
	if RenderGitHub(m) != gh {
		t.Fatal("rendering must be deterministic")
	}
}

// cacheModel is the workspace's model with a remote cache configured.
func cacheModel(remote string, native bool) Model {
	var c Contributions
	c.Add(protocol.CIResult{Credentials: []protocol.Credential{{Env: "NODE_AUTH_TOKEN", Required: true}}})
	c.Add(protocol.CIResult{Credentials: cachecredentials.For(remote)})
	c.Pack, c.Native = true, native
	return NewModel(c, "dev", "main")
}

// stepAfter is what follows a step's first line, up to the next step.
func stepAfter(out, first string) string {
	at := strings.Index(out, first)
	if at < 0 {
		return ""
	}
	rest := out[at+len(first):]
	if end := strings.Index(rest, "\n      - "); end >= 0 {
		rest = rest[:end+1]
	}
	return rest
}

// A remote cache's key and credentials reach every GitHub step that runs
// tasks, from secrets; the audit keeps its verdicts in the cache.
func TestGitHubPassesTheCacheToTheStepsThatRunTasks(t *testing.T) {
	cases := map[string][]string{
		"azblob://acme/smith-cache/main": {"SMITH_CACHE_KEY", "AZURE_STORAGE_SAS_TOKEN"},
		"s3://acme/smith":                {"AWS_ACCESS_KEY_ID", "AWS_REGION", "AWS_SECRET_ACCESS_KEY", "SMITH_CACHE_KEY"},
		"https://cache.acme.dev/smith":   {"SMITH_CACHE_KEY", "SMITH_CACHE_TOKEN"},
	}
	for remote, envs := range cases {
		out := RenderGitHub(cacheModel(remote, true))
		for _, step := range []string{"      - name: Audit\n        run: smith ci audit", "      - name: Verify\n        run: smith ci verify", "      - name: Pack\n", "      - name: Native\n"} {
			body := stepAfter(out, step)
			if body == "" {
				t.Fatalf("%s: no step %q in\n%s", remote, step, out)
			}
			for _, env := range envs {
				if !strings.Contains(body, "          "+env+": ${{ secrets."+env+" }}\n") {
					t.Errorf("%s: %q lacks %s:\n%s", remote, step, env, body)
				}
			}
		}
	}
}

func TestGitHubDoesNotMentionTheCacheWithoutARemote(t *testing.T) {
	out := RenderGitHub(cacheModel("", false))
	if strings.Contains(out, "SMITH_CACHE") || strings.Contains(out, "AZURE_STORAGE") {
		t.Errorf("a workspace without a remote cache gets cache env:\n%s", out)
	}
	if !strings.Contains(out, "      - name: Verify\n        run: smith ci verify\n      # smith:slot after-verify") {
		t.Errorf("verify is unchanged:\n%s", out)
	}
}

func TestTheNativeJobPassesTheCacheWithoutThePublishingSecrets(t *testing.T) {
	remote := "azblob://acme/smith-cache/main"
	out := RenderGitHub(cacheModel(remote, true))
	native := out[strings.Index(out, "\n  native:"):]
	if !strings.Contains(native, "SMITH_CACHE_KEY: ${{ secrets.SMITH_CACHE_KEY }}") || !strings.Contains(native, "AZURE_STORAGE_SAS_TOKEN: ${{ secrets.AZURE_STORAGE_SAS_TOKEN }}") {
		t.Errorf("the native job lacks the cache:\n%s", native)
	}
	if strings.Contains(native, "NODE_AUTH_TOKEN") {
		t.Errorf("a native job that publishes nothing gets no release secret:\n%s", native)
	}
	published := cacheModel(remote, true)
	published.NativePublish = true
	native = RenderGitHub(published)
	native = native[strings.Index(native, "\n  native:"):]
	if strings.Count(native, "SMITH_CACHE_KEY:") != 1 || !strings.Contains(native, "NODE_AUTH_TOKEN: ${{ secrets.NODE_AUTH_TOKEN }}") {
		t.Errorf("a publishing native job gets both, each once:\n%s", native)
	}
}

func TestNoToolchainsWhenNoProjects(t *testing.T) {
	gh := RenderGitHub(NewModel(Contributions{}, "1.2.3", "main"))
	if strings.Contains(gh, "setup-node") || !strings.Contains(gh, "SMITH_VERSION: 1.2.3") || strings.Contains(gh, "env:\n          NODE") {
		t.Fatal(gh)
	}
}

func TestWriteKeepsSlotsAndReportsDrift(t *testing.T) {
	root := t.TempDir()
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"github", "azure"}}}
	m := model()
	changed, _, err := Write(root, cfg, m)
	if err != nil || len(changed) != 2 {
		t.Fatalf("changed %v err %v", changed, err)
	}
	if drift := Drift(root, cfg, m); len(drift) != 0 {
		t.Fatalf("fresh pipelines drift: %v", drift)
	}
	path := filepath.Join(root, ".github", "workflows", "ci.yml")
	raw, _ := os.ReadFile(path)
	custom := strings.Replace(string(raw), "      # smith:slot after-verify\n", "      # smith:slot after-verify\n      - name: Team step\n        run: make e2e\n", 1)
	_ = os.WriteFile(path, []byte(custom), 0o644)
	if drift := Drift(root, cfg, m); len(drift) != 0 {
		t.Fatalf("slot content is not drift: %v", drift)
	}
	m.Toolchains = m.Toolchains[:1] // go only: both pipelines change and the install script is now needed
	if drift := Drift(root, cfg, m); len(drift) != 3 {
		t.Fatalf("changed toolchains must drift: %v", drift)
	}
	if _, _, err := Write(root, cfg, m); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(path)
	if !strings.Contains(string(raw), "run: make e2e") || strings.Contains(string(raw), "setup-node") {
		t.Fatalf("regeneration must keep the slot and drop node:\n%s", raw)
	}
	if changed, _, _ := Write(root, cfg, m); len(changed) != 0 {
		t.Fatalf("an unchanged workspace rewrites nothing: %v", changed)
	}
}

func TestThePreviewIsPostedOnPullRequests(t *testing.T) {
	m := model()
	gh, az := RenderGitHub(m), RenderAzure(m)
	for _, want := range []string{"  pull-requests: write\n", "run: smith release --dry-run --comment\n        env:\n          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n"} {
		if !strings.Contains(gh, want) {
			t.Errorf("GitHub workflow lacks %q", want)
		}
	}
	if !strings.Contains(az, "bash: smith release --dry-run --comment\n    displayName: Release preview\n    condition: and(succeeded(), eq(variables['Build.Reason'], 'PullRequest'))\n    env:\n      SYSTEM_ACCESSTOKEN: $(System.AccessToken)\n") {
		t.Errorf("Azure pipeline lacks the preview comment step:\n%s", az)
	}
}

func TestPackOnlyWhenAProjectPackages(t *testing.T) {
	m := model()
	if strings.Contains(RenderGitHub(m), "smith ci pack") || strings.Contains(RenderAzure(m), "smith ci pack") {
		t.Fatal("no project packages: no pack step")
	}
	m.Pack = true
	gh, az := RenderGitHub(m), RenderAzure(m)
	for _, want := range []string{"run: smith ci pack", "uses: actions/upload-artifact@v7", "name: drop", "path: dist/drop", "if-no-files-found: ignore"} {
		if !strings.Contains(gh, want) {
			t.Errorf("GitHub workflow lacks %q", want)
		}
	}
	for _, want := range []string{"bash: smith ci pack", "publish: dist/drop", "artifact: drop", "refs/heads/main"} {
		if !strings.Contains(az, want) {
			t.Errorf("Azure pipeline lacks %q", want)
		}
	}
	if strings.Index(gh, "smith ci pack") > strings.Index(gh, "smith ci release") {
		t.Error("pack runs before release")
	}
}

func TestGitHubWorkflowMayRequestAnOIDCTokenForNPMProvenance(t *testing.T) {
	m := model()
	if strings.Contains(RenderGitHub(m), "id-token") {
		t.Error("id-token: write without an npm-published app")
	}
	m.Provenance = true
	if !strings.Contains(RenderGitHub(m), "  pull-requests: write\n  id-token: write\n\nconcurrency:") {
		t.Errorf("no id-token permission:\n%s", RenderGitHub(m))
	}
}

func TestGitHubWorkflowAttestsReleaseAssetsOnlyWhenTheyAreAttached(t *testing.T) {
	m := model()
	if gh := RenderGitHub(m); strings.Contains(gh, "attest") {
		t.Errorf("attestation without release assets:\n%s", gh)
	}
	m.Attest = true
	gh := RenderGitHub(m)
	for _, want := range []string{
		"  id-token: write\n  attestations: write\n",
		"uses: actions/attest-build-provenance@v3",
		"subject-path: dist/release/*/*",
		"hashFiles('dist/release/**') != ''",
	} {
		if !strings.Contains(gh, want) {
			t.Errorf("workflow lacks %q:\n%s", want, gh)
		}
	}
	if strings.Index(gh, "smith ci release") > strings.Index(gh, "attest-build-provenance") {
		t.Error("attestation runs before the release attaches its assets")
	}
	m.Provenance = true
	if strings.Count(RenderGitHub(m), "id-token: write") != 1 {
		t.Error("id-token requested twice")
	}
	if az := RenderAzure(m); strings.Contains(az, "attest") {
		t.Error("Azure Pipelines has no attestations")
	}
}

func TestGitHubWorkflowGrantsThePermissionsPluginsAskFor(t *testing.T) {
	var c Contributions
	c.Add(protocol.CIResult{Permissions: []string{"packages: write"}})
	c.Add(protocol.CIResult{Permissions: []string{"packages: write"}})
	m := NewModel(c, "1.0.0", "main")
	if gh := RenderGitHub(m); !strings.Contains(gh, "  pull-requests: write\n  packages: write\n\nconcurrency:") {
		t.Errorf("packages: write once, with the other permissions:\n%s", gh)
	}
	if strings.Contains(RenderAzure(m), "packages") {
		t.Error("Azure Pipelines has no token permissions")
	}
}

func TestPubWorkflowOnlyWhenAnAppPublishesToPubDev(t *testing.T) {
	root := t.TempDir()
	github := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}}}
	m := model()
	if _, ok := Rendered(github, m)[GitHubPubPath]; ok || strings.Contains(RenderGitHub(m), "actions: write") {
		t.Fatal("a pub workflow without a pub.dev app")
	}
	m.Pub = true
	files := Rendered(github, m)
	pub, ok := files[GitHubPubPath]
	if !ok {
		t.Fatal("no pub workflow")
	}
	for _, want := range []string{
		"on:\n  workflow_dispatch:\n    inputs:\n      project:",
		"permissions:\n  contents: read\n  id-token: write\n",
		"uses: dart-lang/setup-dart@v1",
		"run: |\n          npm install --global --registry https://registry.npmjs.org/ monosmith\n          smith --version\n",
		`run: smith release pub --project "$PROJECT" --version "$VERSION"`,
		"PROJECT: ${{ inputs.project }}",
		"# smith:slot before-publish",
	} {
		if !strings.Contains(pub, want) {
			t.Errorf("pub.yml lacks %q:\n%s", want, pub)
		}
	}
	if !strings.Contains(files[GitHubPath], "  pull-requests: write\n  actions: write\n") {
		t.Errorf("ci.yml cannot start the pub workflow:\n%s", files[GitHubPath])
	}
	azure := workspace.Config{CI: workspace.CISection{Providers: []string{"azure"}}}
	if _, ok := Rendered(azure, m)[GitHubPubPath]; ok {
		t.Error("a GitHub workflow for an Azure-only workspace")
	}

	if _, _, err := Write(root, github, m); err != nil {
		t.Fatal(err)
	}
	m.Pub = false
	if changed, _, err := Write(root, github, m); err != nil || !strings.Contains(strings.Join(changed, ","), GitHubPubPath) {
		t.Fatalf("dropping pub: changed %v %v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(root, GitHubPubPath)); !os.IsNotExist(err) {
		t.Error("the pub workflow outlives the pub.dev app")
	}
}

func TestTrustedPublishingSignsInToRubyGemsRightBeforeTheRelease(t *testing.T) {
	var c Contributions
	c.Add(protocol.CIResult{Credentials: []protocol.Credential{{Env: "GEM_HOST_API_KEY", Required: true}, {Env: "PYPI_TOKEN", Required: true}}})
	c.UseRubyGemsTrustedPublishing()
	gh := RenderGitHub(NewModel(c, "dev", "main"))
	signIn := strings.Index(gh, "      - name: Sign in to RubyGems\n        if: github.event_name == 'push' && github.ref_name == 'main'\n        uses: rubygems/configure-rubygems-credentials@v2.1.0\n")
	release := strings.Index(gh, "      - name: Release\n")
	if signIn < 0 || release < 0 || signIn > release {
		t.Fatalf("the sign-in must come right before the release, only on a push to the release branch:\n%s", gh)
	}
	if !strings.Contains(gh, "  id-token: write\n") {
		t.Error("the job may not ask for the OIDC token")
	}
	if strings.Contains(gh, "GEM_HOST_API_KEY") {
		t.Error("the key is still a secret the release reads")
	}
	if !strings.Contains(gh, "PYPI_TOKEN: ${{ secrets.PYPI_TOKEN }}") {
		t.Error("the other credentials went too")
	}
}

func TestWithoutTrustedPublishingTheRubyGemsKeyIsASecret(t *testing.T) {
	var c Contributions
	c.Add(protocol.CIResult{Credentials: []protocol.Credential{{Env: "GEM_HOST_API_KEY", Required: true}}})
	gh := RenderGitHub(NewModel(c, "dev", "main"))
	if strings.Contains(gh, "Sign in to RubyGems") || strings.Contains(gh, "id-token") || !strings.Contains(gh, "GEM_HOST_API_KEY: ${{ secrets.GEM_HOST_API_KEY }}") {
		t.Errorf("GitHub:\n%s", gh)
	}
}

func TestTheNativeJobSignsInToRubyGemsOnlyWhenItPublishes(t *testing.T) {
	var c Contributions
	c.Native = true
	c.UseRubyGemsTrustedPublishing()
	if gh := RenderGitHub(NewModel(c, "dev", "main")); strings.Count(gh, "Sign in to RubyGems") != 1 {
		t.Errorf("a native job that publishes nothing signs in:\n%s", gh)
	}
	c.NativePublish = true
	if gh := RenderGitHub(NewModel(c, "dev", "main")); strings.Count(gh, "Sign in to RubyGems") != 2 {
		t.Errorf("a native job that publishes must sign in too:\n%s", gh)
	}
}

func TestEntraPublishingSignsInToAzureRightBeforeTheRelease(t *testing.T) {
	var c Contributions
	c.Add(protocol.CIResult{Credentials: []protocol.Credential{{Env: "VSCE_PAT", Required: true}, {Env: "PYPI_TOKEN", Required: true}}})
	c.UseVSCEEntra()
	gh := RenderGitHub(NewModel(c, "dev", "main"))
	signIn := strings.Index(gh, "      - name: Sign in to Azure\n        if: github.event_name == 'push' && github.ref_name == 'main'\n        uses: azure/login@v3\n        with:\n          client-id: ${{ secrets.AZURE_CLIENT_ID }}\n          tenant-id: ${{ secrets.AZURE_TENANT_ID }}\n          allow-no-subscriptions: true\n")
	release := strings.Index(gh, "      - name: Release\n")
	if signIn < 0 || release < 0 || signIn > release {
		t.Fatalf("the sign-in must come right before the release, only on a push to the release branch:\n%s", gh)
	}
	if !strings.Contains(gh, "  id-token: write\n") {
		t.Error("the job may not ask for the OIDC token")
	}
	if strings.Contains(gh, "VSCE_PAT") {
		t.Error("the Marketplace token is still a secret the release reads")
	}
	if !strings.Contains(gh, "          VSCE_AZURE_CREDENTIAL: \"true\"\n") {
		t.Errorf("the release is not told to publish with the Azure sign-in:\n%s", gh)
	}
	if !strings.Contains(gh, "PYPI_TOKEN: ${{ secrets.PYPI_TOKEN }}") {
		t.Error("the other credentials went too")
	}
}

func TestEntraPublishingNeedsNothingWhenItIsOff(t *testing.T) {
	var c Contributions
	c.Add(protocol.CIResult{Credentials: []protocol.Credential{{Env: "VSCE_PAT", Required: true}}})
	gh := RenderGitHub(NewModel(c, "dev", "main"))
	if strings.Contains(gh, "Sign in to Azure") || strings.Contains(gh, "id-token") || strings.Contains(gh, "VSCE_AZURE_CREDENTIAL") || !strings.Contains(gh, "VSCE_PAT: ${{ secrets.VSCE_PAT }}") {
		t.Errorf("GitHub:\n%s", gh)
	}
}

func TestTheNativeJobSignsInToAzureOnlyWhenItPublishes(t *testing.T) {
	var c Contributions
	c.Native = true
	c.UseVSCEEntra()
	if gh := RenderGitHub(NewModel(c, "dev", "main")); strings.Count(gh, "Sign in to Azure") != 1 || strings.Count(gh, "VSCE_AZURE_CREDENTIAL") != 1 {
		t.Errorf("a native job that publishes nothing signs in:\n%s", gh)
	}
	c.NativePublish = true
	if gh := RenderGitHub(NewModel(c, "dev", "main")); strings.Count(gh, "Sign in to Azure") != 2 || strings.Count(gh, "VSCE_AZURE_CREDENTIAL") != 2 {
		t.Errorf("a native job that publishes must sign in too:\n%s", gh)
	}
}
