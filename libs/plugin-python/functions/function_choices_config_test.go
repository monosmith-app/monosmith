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
		if c.Axis != "framework" || !c.Plays(protocol.RoleScaffold) || len(c.Needs) != 2 {
			t.Errorf("%s: %+v", host, c)
		}
	}
	if !byName["http-trigger"].Suggested || !byName["python313"].Suggested || byName["python312"].Suggested {
		t.Error("http-trigger and python313 are the suggested answers")
	}
	if got := byName["timer-trigger"].Conflicts; len(got) != 1 || got[0] != "gcp-functions" {
		t.Errorf("Google has no timer trigger: %v", got)
	}
}

func TestPlanOfReadsTheChoices(t *testing.T) {
	plan, ok := PlanOf([]string{"aws-lambda", "queue-trigger", "python312"})
	if !ok || plan != (Plan{Host: "aws-lambda", Trigger: "queue", PythonVersion: "3.12"}) {
		t.Errorf("%+v %v", plan, ok)
	}
	if _, ok := PlanOf([]string{"flask"}); ok {
		t.Error("a project without a host is an ordinary app")
	}
}

func TestDetectHostReadsTheFilesOfTheHost(t *testing.T) {
	for host, files := range map[string]map[string]string{
		"azure-functions": {"function_app.py": "", "host.json": "{}"},
		"aws-lambda":      {"template.yaml": "Transform: AWS::Serverless-2016-10-31\n"},
		"gcp-functions":   {"main.py": "", "requirements.txt": "functions-framework>=3\n"},
		"":                {"main.py": "", "requirements.txt": "flask\n"},
	} {
		dir := t.TempDir()
		for name, body := range files {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if got := DetectHost(dir); got != host {
			t.Errorf("want %q, got %q for %v", host, got, files)
		}
	}
}

func TestAFunctionAppGetsTheTasksOfItsHost(t *testing.T) {
	tasks := map[string]protocol.Task{"start": {}}
	WithFunctionTasks(tasks, "apps/api", "py", []string{"**/*"})
	if tasks["start"].Command[1] != "start.py" || tasks["build"].Outputs[0] != "../../dist/apps/api" || tasks["package"].DependsOn[0] != "build" {
		t.Errorf("%+v", tasks)
	}
}

func TestTheDoctorWarnsOnceForAMissingTool(t *testing.T) {
	ws := t.TempDir()
	for _, p := range []string{"apps/a", "apps/b"} {
		if err := os.MkdirAll(filepath.Join(ws, p), 0o755); err != nil {
			t.Fatal(err)
		}
		for _, f := range []string{"function_app.py", "host.json"} {
			if err := os.WriteFile(filepath.Join(ws, p, f), nil, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	real := lookPath
	defer func() { lookPath = real }()
	lookPath = func(string) (string, error) { return "", os.ErrNotExist }
	got := Findings(ws, []protocol.Project{{Root: "apps/a"}, {Root: "apps/b"}})
	if len(got) != 1 || got[0].OK || !got[0].Warning || got[0].Remedy == "" {
		t.Errorf("%+v", got)
	}
}
