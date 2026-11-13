package scaffolding

import (
	"strings"
	"testing"
)

const csprojTemplate = "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup>\n    <TargetFramework>%s</TargetFramework>\n    <Nullable>enable</Nullable>\n  </PropertyGroup>\n</Project>\n"

func csproj(framework string) string { return strings.Replace(csprojTemplate, "%s", framework, 1) }

func TestShareFrameworkMovesTheMajorityFrameworkToTheProps(t *testing.T) {
	props := "<Project>\n  <PropertyGroup>\n    <Nullable>enable</Nullable>\n  </PropertyGroup>\n</Project>\n"
	projects := map[string]string{"libs/a/A.csproj": csproj("net10.0"), "libs/b/B.csproj": csproj("net10.0"), "libs/c/C.csproj": csproj("net8.0")}
	out, changed, framework := ShareFramework(props, projects)
	if framework != "net10.0" || !strings.Contains(out, "<PropertyGroup>\n    <TargetFramework>net10.0</TargetFramework>\n    <Nullable>") {
		t.Fatalf("props %q framework %q", out, framework)
	}
	if len(changed) != 2 || strings.Contains(changed["libs/a/A.csproj"], "TargetFramework") || !strings.Contains(changed["libs/a/A.csproj"], "<Nullable>") {
		t.Fatalf("changed %v", changed)
	}
}

func TestShareFrameworkKeepsThePropsFrameworkAndCreatesMissingProps(t *testing.T) {
	_, changed, framework := ShareFramework("<Project><PropertyGroup>\n<TargetFramework>net9.0</TargetFramework></PropertyGroup></Project>", map[string]string{"A.csproj": csproj("net9.0"), "B.csproj": csproj("net8.0")})
	if framework != "net9.0" || len(changed) != 1 {
		t.Fatalf("framework %q changed %v", framework, changed)
	}
	if _, changed, framework := ShareFramework("", map[string]string{"A.csproj": csproj("net8.0")}); framework != "" || changed != nil {
		t.Fatalf("one project has nothing to share with: %q %v", framework, changed)
	}
	out, _, framework := ShareFramework("", map[string]string{"A.csproj": csproj("net8.0"), "B.csproj": csproj("net8.0")})
	if framework != "net8.0" || out != "<Project>\n  <PropertyGroup>\n    <TargetFramework>net8.0</TargetFramework>\n  </PropertyGroup>\n</Project>\n" {
		t.Fatalf("new props %q", out)
	}
	if _, changed, framework := ShareFramework("", map[string]string{"A.csproj": "<Project><PropertyGroup><TargetFramework>$(TestTargetFramework)</TargetFramework></PropertyGroup></Project>"}); framework != "" || changed != nil {
		t.Fatalf("a property reference is left alone: %q %v", framework, changed)
	}
	if _, changed, framework := ShareFramework("", map[string]string{"A.csproj": "<Project><PropertyGroup><TargetFrameworks>net8.0;net9.0</TargetFrameworks></PropertyGroup></Project>"}); framework != "" || changed != nil {
		t.Fatalf("multi-targeting is left alone: %q %v", framework, changed)
	}
}
