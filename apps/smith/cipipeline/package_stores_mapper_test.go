package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func storesModel(stores ...protocol.PackageStore) Model {
	return NewModel(Contributions{Stores: stores}, "1.0.0", "main")
}

var (
	npmStore = protocol.PackageStore{Name: "npm", Path: "~/.npm", Env: "NPM_CONFIG_CACHE", Files: []string{"package-lock.json"}}
	goStore  = protocol.PackageStore{Name: "go", Path: "~/go/pkg/mod", Env: "GOMODCACHE", Files: []string{"apps/api/go.sum", "libs/core/go.sum", "libs/extra/go.sum"}}
)

func TestGitHubCachesEachStoreKeyedOnItsLockfiles(t *testing.T) {
	out := RenderGitHub(storesModel(npmStore, goStore))
	for _, want := range []string{
		"      - name: Cache the npm download store\n        uses: actions/cache@v4\n        with:\n          path: ~/.npm\n" +
			"          key: smith-npm-${{ runner.os }}-${{ hashFiles('package-lock.json') }}\n" +
			"          restore-keys: |\n            smith-npm-${{ runner.os }}-\n",
		"          path: ~/go/pkg/mod\n          key: smith-go-${{ runner.os }}-${{ hashFiles('apps/api/go.sum', 'libs/core/go.sum', 'libs/extra/go.sum') }}\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
	if strings.Index(out, "Cache the go download store") > strings.Index(out, "Install smith") {
		t.Errorf("the stores are restored before setup runs:\n%s", out)
	}
}

func TestAzureCachesEachStoreUnderThePipelineWorkspace(t *testing.T) {
	out := RenderAzure(storesModel(npmStore, goStore))
	for _, want := range []string{
		"variables:\n  GOMODCACHE: $(Pipeline.Workspace)/.smith-stores/go\n  NPM_CONFIG_CACHE: $(Pipeline.Workspace)/.smith-stores/npm\n",
		"  - task: Cache@2\n    displayName: Cache the npm download store\n    inputs:\n" +
			"      key: 'smith-npm | \"$(Agent.OS)\" | package-lock.json'\n" +
			"      restoreKeys: |\n        smith-npm | \"$(Agent.OS)\"\n" +
			"      path: $(Pipeline.Workspace)/.smith-stores/npm\n",
		"key: 'smith-go | \"$(Agent.OS)\" | apps/api/go.sum | libs/core/go.sum | libs/extra/go.sum'",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
	if strings.Index(out, "variables:") > strings.Index(out, "pool:") {
		t.Errorf("variables come before the pool:\n%s", out)
	}
}

func TestGitLabCachesEachStoreInsideTheProject(t *testing.T) {
	out := RenderGitLab(storesModel(npmStore, goStore))
	for _, want := range []string{
		"    GOMODCACHE: $CI_PROJECT_DIR/.smith/stores/go\n    NPM_CONFIG_CACHE: $CI_PROJECT_DIR/.smith/stores/npm\n",
		"  cache:\n    - key:\n        prefix: go\n        files: [apps/api/go.sum, libs/core/go.sum]\n      paths:\n        - .smith/stores/go\n",
		"    - key:\n        prefix: npm\n        files: [package-lock.json]\n      paths:\n        - .smith/stores/npm\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestGitLabAllowsFourCachesAndSaysWhatItLeftOut(t *testing.T) {
	var stores []protocol.PackageStore
	for _, name := range []string{"a", "b", "c", "d", "e", "f"} {
		stores = append(stores, protocol.PackageStore{Name: name, Path: "~/." + name, Env: strings.ToUpper(name) + "_CACHE", Files: []string{name + ".lock"}})
	}
	out := RenderGitLab(storesModel(stores...))
	if strings.Count(out, "- key:") != 4 || !strings.Contains(out, "not cached: e, f\n") {
		t.Errorf("four caches and a comment:\n%s", out)
	}
}

func TestNoStoresLeavesThePipelinesAsTheyWere(t *testing.T) {
	m := storesModel()
	for name, out := range map[string]string{"github": RenderGitHub(m), "azure": RenderAzure(m), "gitlab": RenderGitLab(m)} {
		if strings.Contains(out, "download store") || strings.Contains(out, "\nvariables:") || strings.Contains(out, "  cache:") {
			t.Errorf("%s mentions a cache without a store:\n%s", name, out)
		}
	}
}

func TestStoresAreMergedByNameAndSorted(t *testing.T) {
	var c Contributions
	c.Add(protocol.CIResult{Stores: []protocol.PackageStore{npmStore, goStore}})
	c.Add(protocol.CIResult{Stores: []protocol.PackageStore{{Name: "npm", Path: "elsewhere", Env: "X", Files: []string{"x"}}}})
	m := NewModel(c, "1.0.0", "main")
	if len(m.Stores) != 2 || m.Stores[0].Name != "go" || m.Stores[1].Name != "npm" || m.Stores[1].Path != "~/.npm" {
		t.Errorf("%+v", m.Stores)
	}
}
