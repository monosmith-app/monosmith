package engineserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/runevents"
)

// recording is a server whose smith commands are only noted, never run.
func recording(ran *[][]string, stdout string, code int) *Server {
	return &Server{Version: "1", Execute: func(_ context.Context, _ string, args []string, onLine func(runevents.Stream, string)) ([]byte, int, error) {
		*ran = append(*ran, args)
		if onLine != nil {
			onLine(runevents.Stdout, "step")
		}
		return []byte(stdout), code, nil
	}}
}

func TestReleaseIsAStreamedRunThatDoesNotAskAgain(t *testing.T) {
	var ran [][]string
	c := start(t, recording(&ran, "", 0))
	result := c.call("release", map[string]any{"run": "rel-1", "noPush": true})["result"].(map[string]any)
	if result["ok"] != true || result["run"] != "rel-1" {
		t.Fatalf("release answered %v", result)
	}
	want := []string{"run.started", "task.queued release", "task.started release", "task.output release", "task.finished release passed", "run.finished"}
	if fmt.Sprint(c.kinds()) != fmt.Sprint(want) {
		t.Fatalf("events\n got %q\nwant %q", c.kinds(), want)
	}
	if first := c.events[0]; first.Run != "rel-1" || first.Command != "smith release --yes --no-push" {
		t.Fatalf("run.started %+v", first)
	}
	c.call("release", nil)
	if fmt.Sprint(ran) != "[[release --yes --no-push] [release --yes]]" {
		t.Fatalf("commands run %v", ran)
	}
}

func TestAFailedReleaseEndsTheRunFailedWithItsExitCode(t *testing.T) {
	var ran [][]string
	c := start(t, recording(&ran, "", 1))
	result := c.call("release", nil)["result"].(map[string]any)
	if result["ok"] != false || result["exitCode"] != float64(1) {
		t.Fatalf("release answered %v", result)
	}
	if last := c.events[len(c.events)-1]; last.Event != runevents.RunFinished || last.OK == nil || *last.OK {
		t.Fatalf("run.finished %+v", last)
	}
}

func TestConfigSetAndUnsetAreRunsOverTheConfigCommand(t *testing.T) {
	var ran [][]string
	c := start(t, recording(&ran, "", 0))
	c.call("config.set", map[string]any{"setting": "ci.devcontainer", "value": "false"})
	c.call("config.set", map[string]any{"setting": "ci.providers", "values": []string{"github", "azure"}})
	c.call("config.unset", map[string]any{"setting": "release.branch"})
	if fmt.Sprint(ran) != "[[config set -- ci.devcontainer false] [config set -- ci.providers github azure] [config unset release.branch]]" {
		t.Fatalf("commands run %v", ran)
	}
	for _, params := range []map[string]any{{"value": "x"}, {"setting": "--force", "value": "x"}, {"setting": "ci.devcontainer"}} {
		if reply := c.call("config.set", params); reply["error"].(map[string]any)["code"].(float64) != codeInvalidParams {
			t.Errorf("config.set %v answered %v", params, reply)
		}
	}
	if len(ran) != 3 {
		t.Fatalf("a refused request ran %v", ran)
	}
}

func TestConfigListAnswersWithTheJSONOfSmithConfigList(t *testing.T) {
	var ran [][]string
	c := start(t, recording(&ran, `[{"path":"ci.devcontainer","set":false}]`, 0))
	rows := c.call("config.list", nil)["result"].([]any)
	if rows[0].(map[string]any)["path"] != "ci.devcontainer" || fmt.Sprint(ran) != "[[config list --json]]" {
		t.Fatalf("config.list answered %v after %v", rows, ran)
	}
}

func TestConfigGetReadsASettingAndSaysWhetherItIsSet(t *testing.T) {
	t.Setenv("SMITH_WORKSPACE_ROOT", "") // smith sets it for the tasks it runs
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "smith.toml"), []byte("schema = 1\n\n[workspace]\nname = \"acme\"\n\n[ci]\ndevcontainer = false\n"), 0o644)
	c := start(t, &Server{Version: "1", Dir: root})

	set := c.call("config.get", map[string]any{"setting": "ci.devcontainer"})["result"].(map[string]any)
	if set["set"] != true || set["value"] != "false" || set["type"] != "boolean" || set["description"] == "" {
		t.Fatalf("config.get answered %v", set)
	}
	unset := c.call("config.get", map[string]any{"setting": "release.branch"})["result"].(map[string]any)
	if unset["set"] != false || unset["setting"] != "release.branch" {
		t.Fatalf("config.get answered %v", unset)
	}
	if reply := c.call("config.get", map[string]any{"setting": "stacks.web"}); reply["error"].(map[string]any)["code"].(float64) != codeInvalidParams {
		t.Fatalf("an unknown setting answered %v", reply)
	}
}

const teamRecipeFile = `name = "team"
version = "1.2.0"

[choices.language]
ask = "Language"
options = ["python", "csharp"]

[use]
structure = "vertical-slices"

[when.language.python.use]
language = "python"

[when.language.csharp.use]
language = "dotnet"

[when.language.csharp.settings.dotnet]
target_framework = "net10.0"
`

func recipeWorkspace(t *testing.T) *client {
	t.Helper()
	t.Setenv("SMITH_WORKSPACE_ROOT", "") // smith sets it for the tasks it runs
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "smith.toml"), []byte("schema = 1\n\n[workspace]\nname = \"acme\"\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(root, "recipes"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "recipes", "team.recipe.toml"), []byte(teamRecipeFile), 0o644)
	return start(t, &Server{Version: "1", Dir: root})
}

func TestRecipeListFindsTheWorkspacesRecipes(t *testing.T) {
	c := recipeWorkspace(t)
	recipes := c.call("recipe.list", nil)["result"].(map[string]any)["recipes"].([]any)
	if len(recipes) != 1 {
		t.Fatalf("recipe.list answered %v", recipes)
	}
	if r := recipes[0].(map[string]any); r["path"] != "recipes/team.recipe.toml" || r["name"] != "team" || r["version"] != "1.2.0" {
		t.Fatalf("recipe.list answered %v", r)
	}
}

func TestRecipeDescribeShowsWhatARecipeOffersBeforeAnyAnswer(t *testing.T) {
	c := recipeWorkspace(t)
	d := c.call("recipe.describe", map[string]any{"recipe": "recipes/team.recipe.toml"})["result"].(map[string]any)
	if d["name"] != "team" || len(d["choices"].([]any)) != 1 || fmt.Sprint(d["configures"]) != "[dotnet]" {
		t.Fatalf("recipe.describe answered %v", d)
	}
	var language map[string]any
	for _, u := range d["use"].([]any) {
		if u.(map[string]any)["key"] == "language" {
			language = u.(map[string]any)
		}
	}
	if fmt.Sprint(language["alternatives"]) != "map[language=csharp:[dotnet] language=python:[python]]" {
		t.Fatalf("language role %v", language)
	}
}

func TestRecipeRolesSaysWhatTheRecipePutsInEachRoleForTheAnswers(t *testing.T) {
	c := recipeWorkspace(t)
	r := c.call("recipe.roles", map[string]any{"recipe": "recipes/team.recipe.toml", "choices": map[string]string{"language": "csharp"}})["result"].(map[string]any)
	got := map[string]string{}
	for _, role := range r["roles"].([]any) {
		m := role.(map[string]any)
		got[m["key"].(string)] = fmt.Sprint(m["plugins"])
	}
	if fmt.Sprint(got) != "map[language:[dotnet] structure:[vertical-slices]]" {
		t.Fatalf("roles %v", got)
	}
	if fmt.Sprint(r["choices"].([]any)[0].(map[string]any)["answer"]) != "csharp" {
		t.Fatalf("choices %v", r["choices"])
	}
	// A choice without a default has to be answered.
	missing := c.call("recipe.roles", map[string]any{"recipe": "recipes/team.recipe.toml"})
	if msg := missing["error"].(map[string]any)["message"].(string); !strings.Contains(msg, "language") {
		t.Fatalf("an unanswered choice answered %v", missing)
	}
}

func TestARecipeOutsideTheWorkspaceIsRefused(t *testing.T) {
	c := recipeWorkspace(t)
	for _, ref := range []string{"../team.recipe.toml", "--all", "http://example.com/x.recipe.toml"} {
		reply := c.call("recipe.describe", map[string]any{"recipe": ref})
		if reply["error"] == nil {
			t.Errorf("%s answered %v", ref, reply)
		}
	}
}
