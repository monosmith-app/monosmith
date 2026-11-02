package functions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTheChoicesAreValidAndAskTheTriggerAndTheRuntime(t *testing.T) {
	byName := map[string]protocol.Choice{}
	for _, c := range Choices() {
		if err := protocol.ValidateChoice(c); err != nil {
			t.Error(err)
		}
		if c.Default {
			t.Errorf("%s: nothing here is chosen silently", c.Name)
		}
		byName[c.Name] = c
	}
	for _, host := range []string{"azure-functions", "aws-lambda", "gcp-functions"} {
		c := byName[host]
		if c.Axis != "framework" || !c.Plays(protocol.RoleScaffold) || len(c.Needs) != 2 || c.Language != "dotnet" {
			t.Errorf("%s: %+v", host, c)
		}
	}
	if !byName["http-trigger"].Suggested || !byName["dotnet10"].Suggested || byName["dotnet8"].Suggested {
		t.Error("http-trigger and dotnet10 (the current LTS) are the suggested answers")
	}
	if got := byName["dotnet8"].Conflicts; len(got) != 1 || got[0] != "gcp-functions" {
		t.Errorf("Google retires .NET 8 on 2026-11-10: %v", got)
	}
}

func TestPlanOfReadsTheChoices(t *testing.T) {
	plan, ok := PlanOf([]string{"aws-lambda", "queue-trigger", "dotnet8"})
	if !ok || plan != (Plan{Host: "aws-lambda", Trigger: "queue", Runtime: "dotnet8", Framework: "net8.0"}) {
		t.Errorf("%+v %v", plan, ok)
	}
	if _, ok := PlanOf([]string{"react"}); ok {
		t.Error("a project without a host is an ordinary app")
	}
}

func TestDetectHostReadsTheFilesOfTheHost(t *testing.T) {
	for host, tc := range map[string]struct {
		csproj string
		files  map[string]string
	}{
		"azure-functions": {`<PackageReference Include="Microsoft.Azure.Functions.Worker" />`, map[string]string{"host.json": "{}"}},
		"aws-lambda":      {`<PackageReference Include="Amazon.Lambda.Core" />`, map[string]string{"template.yaml": "Transform: AWS::Serverless-2016-10-31\n"}},
		"gcp-functions":   {`<PackageReference Include="Google.Cloud.Functions.Hosting" />`, nil},
		"":                {`<Project Sdk="Microsoft.NET.Sdk" />`, nil},
	} {
		dir := t.TempDir()
		for name, body := range tc.files {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if got := DetectHost(dir, tc.csproj); got != host {
			t.Errorf("want %q, got %q", host, got)
		}
	}
}

func TestTheDoctorWarnsOnceForAMissingTool(t *testing.T) {
	ws := t.TempDir()
	for _, p := range []string{"apps/a", "apps/b"} {
		if err := os.MkdirAll(filepath.Join(ws, p), 0o755); err != nil {
			t.Fatal(err)
		}
		_ = os.WriteFile(filepath.Join(ws, p, "host.json"), nil, 0o644)
		_ = os.WriteFile(filepath.Join(ws, p, "X.csproj"), []byte(`<PackageReference Include="Microsoft.Azure.Functions.Worker" />`), 0o644)
	}
	real := lookPath
	defer func() { lookPath = real }()
	lookPath = func(string) (string, error) { return "", os.ErrNotExist }
	got := Findings(ws, []protocol.Project{{Root: "apps/a"}, {Root: "apps/b"}})
	if len(got) != 1 || got[0].OK || !got[0].Warning || got[0].Remedy == "" {
		t.Errorf("%+v", got)
	}
}
