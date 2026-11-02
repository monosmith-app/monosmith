package guide

import (
	"reflect"
	"testing"
)

func formTitled(t *testing.T, title string) Action {
	t.Helper()
	for _, a := range parityForms() {
		if a.Title == title {
			return a
		}
	}
	t.Fatalf("no form titled %q", title)
	return Action{}
}

func TestParityFormsBuildTheCommandOfTheirAnswers(t *testing.T) {
	tests := []struct {
		title   string
		answers map[string][]string
		want    []string
	}{
		{EnvironmentsAddTitle, map[string][]string{"project": {"web"}, "names": {"qa, stage"}}, []string{"environments", "add", "web", "qa", "stage"}},
		{EnvironmentsRemoveTitle, map[string][]string{"project": {"web"}, "names": {"dev"}}, []string{"environments", "remove", "web", "dev"}},
		{"Add a ready-made pipeline", map[string][]string{"template": {"deploy-pages"}, "project": {"web"}}, []string{"pipeline", "deploy-pages", "--project", "web"}},
		{"Add a ready-made pipeline", map[string][]string{"template": {"e2e"}}, []string{"pipeline", "e2e"}},
		{"Make the icons of a Tauri app", map[string][]string{"logo": {"logo.png"}}, []string{"tauri", "icon", "logo.png"}},
		{"Make the icons of a Tauri app", map[string][]string{"logo": {"logo.png"}, "project": {"desktop"}}, []string{"tauri", "icon", "logo.png", "-p", "desktop"}},
		{"Set the Docker Hub descriptions", map[string][]string{"mode": {"preview"}}, []string{"docker", "description", "--dry-run"}},
		{"Set the Docker Hub descriptions", map[string][]string{"mode": {"set"}}, []string{"docker", "description"}},
		{"Why would a task run again?", map[string][]string{"task": {"web:build"}}, []string{"why-miss", "web:build"}},
		{"Connect smith to a coding assistant", map[string][]string{"tools": {"--claude", "--vscode"}}, []string{"mcp", "install", "--claude", "--vscode"}},
	}
	for _, tt := range tests {
		if got := formTitled(t, tt.title).Argv(tt.answers); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: got %v, want %v", tt.title, got, tt.want)
		}
	}
}

func TestThePipelineFormAsksTheAppOnlyForADeployTemplate(t *testing.T) {
	steps := formTitled(t, "Add a ready-made pipeline").Steps
	if !steps[1].Skip(map[string][]string{"template": {"e2e"}}) {
		t.Error("e2e needs no app, but the form asks for one")
	}
	if steps[1].Skip(map[string][]string{"template": {"deploy-pages"}}) {
		t.Error("deploy-pages builds an app, but the form does not ask which")
	}
}

func TestTheNamesOfAnAppsOwnEnvironmentsStartOnTheWorkspacesNamesAndCanBeEdited(t *testing.T) {
	var step Step
	for _, s := range webEnvironmentSteps(nil) {
		if s.Key == envNamesKey {
			step = s
		}
	}
	if got := step.Initial(nil); got != "lcl, dev, uat, prd" {
		t.Errorf("without a workspace list the names start on the default four, got %q", got)
	}
	action := WithWorkspaceEnvironments(Action{Steps: []Step{step}}, []string{"qa", "prd"})
	if got := action.Steps[0].Initial(nil); got != "qa, prd" {
		t.Errorf("the names start on the workspace's, got %q", got)
	}
	shown := action.Steps[0]
	shown.Skip = nil // the app is a web app: the question is asked
	m := NewModel([]Action{{Title: "Add", Steps: []Step{shown}}})
	m, _ = m.Pick("Add")
	if got := m.input.Value(); got != "qa, prd" {
		t.Errorf("the text field holds %q, want the workspace names to edit", got)
	}
}

func TestAProjectPickReplacesTheTextStepOfTheSameKey(t *testing.T) {
	action := WithProjectPick(environmentsForm("add"), "project", []string{"web", "site"})
	if got := len(action.Steps[0].Options); got != 2 {
		t.Fatalf("the project step lists %d projects, want 2", got)
	}
	if action.Steps[0].Prompt != "Which web app?" {
		t.Errorf("the pick lost the prompt of the step it replaced: %q", action.Steps[0].Prompt)
	}
}
