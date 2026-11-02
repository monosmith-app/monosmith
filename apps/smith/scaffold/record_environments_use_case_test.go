package scaffold

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
)

func TestEnvironmentsTomlRecordsTheAppAndTheWorkspaceList(t *testing.T) {
	src := "[workspace]\nname = \"acme\"\n"
	got := EnvironmentsToml(src, "web", webenv.Settings{Names: webenv.DefaultNames, Output: webenv.OutputProject})
	for _, want := range []string{
		"environments = [\"lcl\", \"dev\", \"uat\", \"prd\"]",
		"[projects.web.settings.vite]",
		"default_environment = \"prd\"",
		"output = \"project\"",
		"env_files = \"commit\"",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("smith.toml misses %q:\n%s", want, got)
		}
	}
	cfg, err := workspace.Parse([]byte(got))
	if err != nil {
		t.Fatalf("smith.toml still loads: %v\n%s", err, got)
	}
	if strings.Join(cfg.Workspace.Environments, ",") != "lcl,dev,uat,prd" {
		t.Errorf("workspace list: %v", cfg.Workspace.Environments)
	}
	if _, ok := webenv.FromMap(cfg.Projects["web"].Settings["vite"]); !ok {
		t.Errorf("the app's table reads back: %+v", cfg.Projects["web"].Settings)
	}
}

func TestEnvironmentsTomlKeepsAWorkspaceListThatIsThere(t *testing.T) {
	src := "[workspace]\nenvironments = [\"dev\", \"prd\"]\n"
	got := EnvironmentsToml(src, "web", webenv.Settings{Names: []string{"lcl", "prd"}})
	if !strings.Contains(got, "[workspace]\nenvironments = [\"dev\", \"prd\"]") {
		t.Errorf("the workspace list is the owner's:\n%s", got)
	}
}

func TestEnvironmentLinesListEveryFileAndTask(t *testing.T) {
	lines := EnvironmentLines("apps/web/", webenv.Settings{Names: []string{"dev", "prd"}})
	var text strings.Builder
	for _, l := range lines {
		text.WriteString(l.Path + " " + l.Note + "\n")
	}
	for _, want := range []string{"apps/web/.env.dev", "apps/web/.env.prd", "apps/web/vite.build.ts", "build:dev -> dist/apps/web/dev", "build -> dist/apps/web/prd"} {
		if !strings.Contains(text.String(), want) {
			t.Errorf("the plan misses %q:\n%s", want, text.String())
		}
	}
}

func TestTheEnvironmentQuestionsMarkTheirDefaultAndTakeEnter(t *testing.T) {
	var out bytes.Buffer
	got := askEnvironmentChoices(webenv.Settings{Names: webenv.DefaultNames}, bufio.NewReader(strings.NewReader("\n\n\n")), &out)
	if got.Default != "prd" || got.Output != webenv.OutputWorkspace || got.EnvFiles != webenv.FilesCommit {
		t.Errorf("Enter takes the defaults: %+v", got)
	}
	for _, want := range []string{"(default)", "dist/apps/<app>/<name>", "hold no secret"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the questions miss %q:\n%s", want, out.String())
		}
	}
	got = askEnvironmentChoices(webenv.Settings{Names: webenv.DefaultNames}, bufio.NewReader(strings.NewReader("uat\n2\n2\n")), &out)
	if got.Default != "uat" || got.Output != webenv.OutputProject || got.EnvFiles != webenv.FilesExample {
		t.Errorf("answers by name and number: %+v", got)
	}
	given := askEnvironmentChoices(webenv.Settings{Names: webenv.DefaultNames, Default: "dev", Output: "project", EnvFiles: "example"}, bufio.NewReader(strings.NewReader("")), &out)
	if given.Default != "dev" || given.Output != "project" || given.EnvFiles != "example" {
		t.Errorf("a flag is never asked again: %+v", given)
	}
}
