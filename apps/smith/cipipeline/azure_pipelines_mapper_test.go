package cipipeline

import (
	"strings"
	"testing"
)

// azureStepAfter is what follows a step's first line, up to the next step.
func azureStepAfter(out, first string) string {
	at := strings.Index(out, first)
	if at < 0 {
		return ""
	}
	rest := out[at+len(first):]
	if end := strings.Index(rest, "\n  - "); end >= 0 {
		rest = rest[:end+1]
	}
	if end := strings.Index(rest, "\n      - "); end >= 0 {
		rest = rest[:end+1]
	}
	return rest
}

// Azure Pipelines exposes a secret variable to a step only through env:, so
// each step that runs tasks maps the cache's key and credentials itself.
func TestAzureMapsTheCacheSecretsIntoEveryStepThatRunsTasks(t *testing.T) {
	cases := map[string][]string{
		"azblob://acme/smith-cache/main": {"AZURE_STORAGE_SAS_TOKEN", "SMITH_CACHE_KEY"},
		"s3://acme/smith":                {"AWS_ACCESS_KEY_ID", "AWS_REGION", "AWS_SECRET_ACCESS_KEY", "SMITH_CACHE_KEY"},
		"https://cache.acme.dev/smith":   {"SMITH_CACHE_KEY", "SMITH_CACHE_TOKEN"},
	}
	for remote, envs := range cases {
		out := RenderAzure(cacheModel(remote, true))
		for _, step := range []string{"- bash: smith ci audit\n", "- bash: smith ci verify\n", "- bash: smith ci pack\n", "- bash: smith ci native\n"} {
			body := azureStepAfter(out, step)
			if body == "" {
				t.Fatalf("%s: no step %q in\n%s", remote, step, out)
			}
			if !strings.Contains(body, "env:\n") {
				t.Errorf("%s: %q has no env block:\n%s", remote, step, body)
			}
			for _, env := range envs {
				if !strings.Contains(body, " "+env+": $("+env+")\n") {
					t.Errorf("%s: %q does not map %s:\n%s", remote, step, env, body)
				}
			}
		}
	}
}

func TestAzureDoesNotMentionTheCacheWithoutARemote(t *testing.T) {
	out := RenderAzure(cacheModel("", false))
	if strings.Contains(out, "SMITH_CACHE") {
		t.Errorf("a workspace without a remote cache gets cache env:\n%s", out)
	}
}

func TestANativeProjectAddsAJobPerOS(t *testing.T) {
	m := model()
	if strings.Contains(RenderGitHub(m), "native:") || strings.Contains(RenderAzure(m), "jobs:") {
		t.Fatal("no native project: no native job, and Azure keeps its single job")
	}
	m.Native, m.NativeRunners = true, DefaultNativeRunners
	gh, az := RenderGitHub(m), RenderAzure(m)
	for _, want := range []string{
		"\n  native:\n    needs: ci\n    strategy:\n      fail-fast: false\n      matrix:\n        os: [ubuntu-latest, windows-latest, macos-latest]\n    runs-on: ${{ matrix.os }}\n",
		"      - uses: actions/setup-node@v7\n",
		"      - name: Install smith\n        shell: bash\n",
		"      - name: Native\n        run: smith ci native\n        env:\n          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n",
		"          name: native-${{ matrix.os }}\n",
	} {
		if !strings.Contains(gh, want) {
			t.Errorf("GitHub workflow lacks %q", want)
		}
	}
	for _, want := range []string{
		"\njobs:\n  - job: ci\n    pool:\n      vmImage: ubuntu-latest\n    steps:\n      - checkout: self\n",
		"      # smith:slot before-setup\n",
		"      - bash: smith ci verify\n",
		"  - job: native\n    dependsOn: ci\n    strategy:\n      matrix:\n        ubuntu_latest:\n          image: ubuntu-latest\n        windows_latest:\n          image: windows-latest\n",
		"    pool:\n      vmImage: $(image)\n",
		"      - task: UseNode@1\n",
		"      - bash: smith ci native\n",
		"        artifact: native-$(image)\n",
	} {
		if !strings.Contains(az, want) {
			t.Errorf("Azure pipeline lacks %q", want)
		}
	}
	if strings.Contains(az, "\nsteps:\n") || strings.Contains(az, "\npool:\n") {
		t.Error("with jobs, Azure has no top-level steps or pool")
	}
	if native := gh[strings.Index(gh, "\n  native:\n"):]; strings.Contains(native, "NODE_AUTH_TOKEN") {
		t.Error("a native job that publishes nothing gets no publishing secret")
	}
}

func TestANativePackagePublishesFromTheNativeJob(t *testing.T) {
	m := model()
	m.Native, m.NativePublish, m.NativeRunners = true, true, DefaultNativeRunners
	gh := RenderGitHub(m)
	native := gh[strings.Index(gh, "\n  native:\n"):]
	step := native[strings.Index(native, "      - name: Native\n"):strings.Index(native, "      - uses: actions/upload-artifact@v7\n")]
	for _, want := range []string{"GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n", "NODE_AUTH_TOKEN: ${{ secrets.NODE_AUTH_TOKEN }}\n", "PYPI_TOKEN: ${{ secrets.PYPI_TOKEN }}\n"} {
		if !strings.Contains(step, want) {
			t.Errorf("each native job publishes its platform, with the release credentials; lacks %q:\n%s", want, step)
		}
	}
	if strings.Count(step, "GITHUB_TOKEN:") != 1 {
		t.Errorf("GITHUB_TOKEN once:\n%s", step)
	}
	az := RenderAzure(m)
	job := az[strings.Index(az, "  - job: native\n"):]
	azStep := job[strings.Index(job, "      - bash: smith ci native\n"):strings.Index(job, "      - publish: dist/drop\n")]
	if !strings.Contains(azStep, "        env:\n") || !strings.Contains(azStep, "          NODE_AUTH_TOKEN: $(NODE_AUTH_TOKEN)\n") {
		t.Errorf("the Azure native job publishes too, with the release credentials:\n%s", azStep)
	}
	m.NativePublish = false
	if az := RenderAzure(m); strings.Contains(az[strings.Index(az, "  - job: native\n"):], "NODE_AUTH_TOKEN") {
		t.Error("an Azure native job that publishes nothing gets no secret")
	}
}

// A team step kept in a slot follows the slot when the Azure pipeline nests
// its steps under jobs.
func TestSlotsAreReindentedWhenTheyMove(t *testing.T) {
	m := model()
	old := strings.Replace(RenderAzure(m), "  # smith:slot before-setup\n", "  # smith:slot before-setup\n  - bash: make tools\n    displayName: Team tools\n", 1)
	m.Native, m.NativeRunners = true, DefaultNativeRunners
	kept, lost := KeepSlots(old, RenderAzure(m))
	if len(lost) != 0 || !strings.Contains(kept, "      # smith:slot before-setup\n      - bash: make tools\n        displayName: Team tools\n      # smith:slot-end before-setup") {
		t.Fatalf("lost %v:\n%s", lost, kept)
	}
}
