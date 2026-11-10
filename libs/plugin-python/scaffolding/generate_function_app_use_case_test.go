package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerateWritesAFunctionAppForEveryHostAndTrigger(t *testing.T) {
	venv.Runner = func(_ string, _ []string) error { return nil }
	for _, host := range []string{"azure-functions", "aws-lambda", "gcp-functions"} {
		for _, trigger := range []string{"http-trigger", "timer-trigger", "queue-trigger", "blob-trigger"} {
			if host == "gcp-functions" && trigger == "timer-trigger" {
				continue
			}
			t.Run(host+"/"+trigger, func(t *testing.T) {
				ws := t.TempDir()
				py := venv.Path(ws)
				_ = os.MkdirAll(filepath.Dir(py), 0o755)
				_ = os.WriteFile(py, nil, 0o755)
				res, err := Generate(ws, protocol.GenerateParams{
					Kind: protocol.ProjectApp, Name: "api", Directory: "apps/api", Stack: "s",
					Choices: []string{host, trigger, "python313"}, Workspace: protocol.WorkspaceInfo{Name: "acme"},
				})
				if err != nil {
					t.Fatal(err)
				}
				read := func(rel string) string {
					raw, err := os.ReadFile(filepath.Join(ws, "apps", "api", filepath.FromSlash(rel)))
					if err != nil {
						t.Fatalf("%s: %v (wrote %v)", rel, err, res.Files)
					}
					return string(raw)
				}
				for _, rel := range []string{"src/acme_api/functions/greeting_handler.py", "src/acme_api/functions/__init__.py", "tests/test_greeting_handler.py", "package.py", "start.py", "README.md"} {
					read(rel)
				}
				if _, err := os.Stat(filepath.Join(ws, "apps", "api", "src", "acme_api", "__main__.py")); err == nil {
					t.Error("a function app has no __main__.py")
				}
				pyproject := read("pyproject.toml")
				if !strings.Contains(pyproject, `requires-python = ">=3.12"`) || !strings.Contains(pyproject, `kind = "app"`) {
					t.Errorf("pyproject:\n%s", pyproject)
				}
				switch host {
				case "azure-functions":
					for _, rel := range []string{"function_app.py", "host.json", "requirements.txt", "local.settings.sample.json"} {
						read(rel)
					}
					if !strings.Contains(read(".gitignore"), "local.settings.json") || !strings.Contains(pyproject, "azure-functions>=") {
						t.Error("the settings with secrets are ignored and the library is a dependency")
					}
					if _, err := os.Stat(filepath.Join(ws, "apps", "api", "local.settings.json")); err == nil {
						t.Error("only the sample is written")
					}
				case "aws-lambda":
					template := read("template.yaml")
					for _, want := range []string{"Runtime: python3.13", "Handler: acme_api.functions.greeting_handler.lambda_handler", "CodeUri: ../../dist/apps/api/app"} {
						if !strings.Contains(template, want) {
							t.Errorf("template.yaml lacks %q:\n%s", want, template)
						}
					}
					read("events/greeting.json")
				case "gcp-functions":
					if !strings.Contains(read("main.py"), "from src.acme_api.functions import") || !strings.Contains(read("requirements.txt"), "functions-framework") {
						t.Error("the Functions Framework loads main.py")
					}
				}
				joined := strings.Join(res.Notes, "\n")
				if !strings.Contains(joined, "smith run start") {
					t.Errorf("next steps: %v", res.Notes)
				}
				if host == "azure-functions" && !strings.Contains(joined, "smith pipeline deploy-azure-function") {
					t.Errorf("the Azure text names the deploy pipeline: %v", res.Notes)
				}
			})
		}
	}
}

func TestGenerateRefusesAFunctionAppThatIsNotAnApp(t *testing.T) {
	_, err := Generate(t.TempDir(), protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Choices: []string{"azure-functions"}})
	if err == nil || !strings.Contains(err.Error(), "an app") {
		t.Errorf("got %v", err)
	}
}
