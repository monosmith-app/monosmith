package scaffolding

import (
	"encoding/xml"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// The C# templates cannot be built on a machine without the .NET SDK, so these
// tests check what can be checked without one: the files, that every project file
// is well-formed XML, the packages and attributes that each host needs, and the
// tasks that discovery finds.
func TestGenerateWritesAFunctionAppForEveryHostAndTrigger(t *testing.T) {
	for _, host := range []string{"azure-functions", "aws-lambda", "gcp-functions"} {
		for _, trigger := range []string{"http-trigger", "timer-trigger", "queue-trigger", "blob-trigger"} {
			if host == "gcp-functions" && trigger == "timer-trigger" {
				continue
			}
			t.Run(host+"/"+trigger, func(t *testing.T) {
				ws := t.TempDir()
				res, err := Generate(ws, protocol.GenerateParams{
					Kind: protocol.ProjectApp, Name: "api", Directory: "apps/api", Stack: "s",
					Choices: []string{host, trigger, "dotnet10"}, Workspace: protocol.WorkspaceInfo{Name: "acme"},
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
				// Every project file is well-formed XML.
				_ = filepath.WalkDir(ws, func(path string, d fs.DirEntry, _ error) error {
					if d != nil && strings.HasSuffix(path, ".csproj") {
						raw, _ := os.ReadFile(path)
						var doc struct{ XMLName xml.Name }
						if err := xml.Unmarshal(raw, &doc); err != nil || doc.XMLName.Local != "Project" {
							t.Errorf("%s is not a well-formed project file: %v", path, err)
						}
					}
					return nil
				})
				csproj := read("Acme.Api.csproj")
				handler := read("Functions/GreetingHandler.cs")
				read("tests/Functions/GreetingHandlerTests.cs")
				read("README.md")
				if _, err := os.Stat(filepath.Join(ws, "apps", "api", "Greeting", "GreetUseCase.cs")); err != nil {
					t.Error("the greeting slice is written")
				}
				for _, want := range []string{"<TargetFramework>net10.0</TargetFramework>", "<SmithKind>app</SmithKind>", "<SmithStack>s</SmithStack>", "<Compile Remove=\"tests/**\" />"} {
					if !strings.Contains(csproj, want) {
						t.Errorf("csproj lacks %q:\n%s", want, csproj)
					}
				}
				switch host {
				case "azure-functions":
					for _, want := range []string{"Microsoft.Azure.Functions.Worker\" Version=", "Microsoft.Azure.Functions.Worker.Sdk", "<OutputType>Exe</OutputType>", "<AzureFunctionsVersion>v4</AzureFunctionsVersion>"} {
						if !strings.Contains(csproj, want) {
							t.Errorf("csproj lacks %q", want)
						}
					}
					for _, rel := range []string{"Program.cs", "host.json", "local.settings.sample.json"} {
						read(rel)
					}
					if !strings.Contains(read(".gitignore"), "local.settings.json") || !strings.Contains(read("Program.cs"), "ConfigureFunctionsWebApplication") || !strings.Contains(handler, "[Function(") {
						t.Error("the settings with secrets are ignored, and the worker is configured")
					}
					if _, err := os.Stat(filepath.Join(ws, "apps", "api", "local.settings.json")); err == nil {
						t.Error("only the sample is written")
					}
				case "aws-lambda":
					if strings.Contains(csproj, "<OutputType>Exe</OutputType>") || !strings.Contains(csproj, "Amazon.Lambda.Core") || !strings.Contains(handler, "[assembly: LambdaSerializer(") {
						t.Errorf("a Lambda project is a class library with the serializer:\n%s\n%s", csproj, handler)
					}
					template := read("template.yaml")
					for _, want := range []string{"Runtime: dotnet10", "CodeUri: ../../dist/dotnet/publish/Acme.Api/release", "Handler: Acme.Api::Acme.Api.Functions.GreetingHandler::Handle"} {
						if !strings.Contains(template, want) {
							t.Errorf("template.yaml lacks %q:\n%s", want, template)
						}
					}
					read("Events/greeting.json")
				case "gcp-functions":
					if !strings.Contains(csproj, "Google.Cloud.Functions.Hosting") || !strings.Contains(csproj, "<OutputType>Exe</OutputType>") || !strings.Contains(handler, "Function") {
						t.Errorf("csproj:\n%s", csproj)
					}
				}
				// Discovery finds an app whose tasks are those of its host.
				found, err := discovery.Discover(ws, []string{"apps/api/Acme.Api.csproj", "apps/api/tests/Acme.Api.Tests.csproj"})
				if err != nil || len(found.Projects) != 1 {
					t.Fatalf("%v %+v", err, found.Projects)
				}
				tasks := found.Projects[0].Tasks
				if found.Projects[0].Kind != protocol.ProjectApp || tasks["package"].Command[1] != "publish" || tasks["test"].Command[2] != "tests/Acme.Api.Tests.csproj" {
					t.Errorf("tasks %+v", tasks)
				}
				switch host {
				case "azure-functions":
					if tasks["start"].Command[0] != "func" || tasks["start"].Cwd != "dist/dotnet/publish/Acme.Api/release" {
						t.Errorf("start %+v", tasks["start"])
					}
				case "aws-lambda":
					want := "invoke"
					if trigger == "http-trigger" {
						want = "start-api"
					}
					if tasks["start"].Command[0] != "sam" || tasks["start"].Command[2] != want {
						t.Errorf("start %+v", tasks["start"])
					}
				case "gcp-functions":
					if tasks["start"].Command[0] != "dotnet" {
						t.Errorf("Google's function is its own host: %+v", tasks["start"])
					}
				}
				joined := strings.Join(res.Notes, "\n")
				if !strings.Contains(joined, "smith run start") || (host == "azure-functions") != strings.Contains(joined, "smith pipeline deploy-azure-function") {
					t.Errorf("next steps: %v", res.Notes)
				}
			})
		}
	}
}

func TestADotnet8RuntimeSetsTheFrameworkOfTheAppOnly(t *testing.T) {
	ws := t.TempDir()
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "api", Directory: "apps/api", Choices: []string{"aws-lambda", "http-trigger", "dotnet8"}, Workspace: protocol.WorkspaceInfo{Name: "acme"}}); err != nil {
		t.Fatal(err)
	}
	app, _ := os.ReadFile(filepath.Join(ws, "apps", "api", "Acme.Api.csproj"))
	template, _ := os.ReadFile(filepath.Join(ws, "apps", "api", "template.yaml"))
	if !strings.Contains(string(app), "<TargetFramework>net8.0</TargetFramework>") || !strings.Contains(string(template), "Runtime: dotnet8") {
		t.Errorf("%s\n%s", app, template)
	}
}

func TestGenerateRefusesAFunctionAppThatIsNotAnApp(t *testing.T) {
	_, err := Generate(t.TempDir(), protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Choices: []string{"azure-functions"}})
	if err == nil || !strings.Contains(err.Error(), "an app") {
		t.Errorf("got %v", err)
	}
}
