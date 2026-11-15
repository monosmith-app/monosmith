package layout

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// project writes files with their contents under root/at and returns that folder.
func project(t *testing.T, at string, files map[string]string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), filepath.FromSlash(at))
	for name, text := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func goFile(imports ...string) string {
	return "package x\n\nimport (\n\t\"" + strings.Join(imports, "\"\n\t\"") + "\"\n)\n"
}

func TestAGoAppKeepsItsSlicesInZonesAndItsImportsInDirection(t *testing.T) {
	dir := project(t, "apps/api", map[string]string{
		"go.mod":  "module ex.com/api\n",
		"main.go": goFile("ex.com/api/features/orders"),
		// A use case calls an integration; a vendor package belongs in the integration.
		"features/orders/place_order_use_case.go": goFile("ex.com/api/integrations/crm", "github.com/lib/pq/oid"),
		"features/orders/order_policy.go":         goFile("ex.com/api/integrations/crm"),
		"features/orders/order_model.go":          goFile("ex.com/api/shared/text"),
		"features/orders/order_policy_test.go":    goFile("ex.com/api/integrations/crm", "github.com/lib/pq"),
		"integrations/crm/crm_client.go":          goFile("github.com/lib/pq", "ex.com/api/shared/text"),
		"shared/text/text_mapper.go":              goFile("ex.com/api/features/orders"),
		"reports/report_use_case.go":              goFile(),
	})
	r, err := Check(dir, "go", []string{"github.com/lib/pq"})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		"reports/ - an app keeps its slices in features/, integrations/ or shared/",
		`features/orders/place_order_use_case.go - "github.com/lib/pq/oid" is a vendor package`,
		"features/orders/order_policy.go - a policy holds no I/O, so it imports no integration",
		"shared/text/text_mapper.go - shared/ holds no business code, so it never imports a feature",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 4 {
		t.Errorf("want 4 problems (shared/ is a zone here, a test may use vendors and integrations), got:\n%s", got)
	}
}

func TestAGoLibraryWithoutZonesKeepsItsSlicesAtTheRoot(t *testing.T) {
	dir := project(t, "libs/money", map[string]string{
		"go.mod":                        "module ex.com/money\n",
		"doc.go":                        "package money\n",
		"billing/fee_policy.go":         goFile(),
		"billing/doc.go":                "package billing\n",
		"billing/shared/rate_mapper.go": goFile(),
	})
	r, err := Check(dir, "go", nil)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	if strings.Contains(got, "billing/ -") {
		t.Errorf("a library without zones keeps its slices at the root:\n%s", got)
	}
	if !strings.Contains(got, `shared/ - "shared" is a junk drawer`) {
		t.Errorf("shared outside the zones of a zoned project is a junk drawer:\n%s", got)
	}
}

func TestAGoLibraryWithOneZoneFolderIsHeldToTheZones(t *testing.T) {
	dir := project(t, "libs/office", map[string]string{
		"go.mod":                             "module ex.com/office\n",
		"integrations/excel/excel_client.go": goFile(),
		"sheets/sheet_model.go":              goFile(),
	})
	r, err := Check(dir, "go", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(r.Problems, "\n"); !strings.Contains(got, "sheets/ - an app keeps its slices") {
		t.Errorf("a library with one zone folder keeps every slice in a zone:\n%s", got)
	}
}

func TestAKotlinAppHasItsZonesUnderItsPackagePath(t *testing.T) {
	base := "src/main/kotlin/dev/ex/ide/"
	dir := project(t, "apps/ide", map[string]string{
		"build.gradle.kts":                         "",
		base + "features/tasks/RunTaskUseCase.kt":  "package dev.ex.ide.features.tasks\n\nimport dev.ex.ide.integrations.smith.SmithClient\n",
		base + "features/tasks/TaskPolicy.kt":      "package dev.ex.ide.features.tasks\n\nimport dev.ex.ide.integrations.smith.SmithClient\n",
		base + "integrations/smith/SmithClient.kt": "package dev.ex.ide.integrations.smith\n",
		base + "shared/words/WordsMapper.kt":       "package dev.ex.ide.shared.words\n\nimport dev.ex.ide.features.tasks.RunTaskUseCase\n",
		base + "toolwindow/ToolWindowHandler.kt":   "package dev.ex.ide.toolwindow\n",
	})
	r, err := Check(dir, "jvm", nil)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		base + "toolwindow/ - an app keeps its slices",
		base + "features/tasks/TaskPolicy.kt - a policy holds no I/O",
		base + "shared/words/WordsMapper.kt - shared/ holds no business code",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "RunTaskUseCase.kt -") {
		t.Errorf("a use case may call an integration:\n%s", got)
	}
}
