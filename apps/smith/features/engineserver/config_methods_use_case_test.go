package engineserver

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigListAnswersWithTheJSONOfSmithConfigList(t *testing.T) {
	var ran [][]string
	c := start(t, recording(&ran, `[{"path":"ci.devcontainer","set":false}]`, 0))
	rows := c.call("config.list", nil)["result"].([]any)
	if rows[0].(map[string]any)["path"] != "ci.devcontainer" || fmt.Sprint(ran) != "[[config list --json]]" {
		t.Fatalf("config.list answered %v after %v", rows, ran)
	}
}

func TestConfigListOfAProjectAsksForThatProjectsView(t *testing.T) {
	var ran [][]string
	c := start(t, recording(&ran, `{"name":"web","manifest":"apps/web/package.json","facts":[],"settings":[]}`, 0))
	view := c.call("config.list", map[string]any{"project": "web"})["result"].(map[string]any)
	if view["manifest"] != "apps/web/package.json" || fmt.Sprint(ran) != "[[config list --json -p web]]" {
		t.Fatalf("config.list answered %v after %v", view, ran)
	}
	if reply := c.call("config.list", map[string]any{"project": "--all"}); reply["error"].(map[string]any)["code"].(float64) != codeInvalidParams {
		t.Fatalf("a project that is an option answered %v", reply)
	}
}

func TestAProjectSettingIsReadAtItsPath(t *testing.T) {
	t.Setenv("SMITH_WORKSPACE_ROOT", "")
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "smith.toml"), []byte("schema = 1\n\n[workspace]\nname = \"acme\"\n\n[projects.web]\ntag_prefix = \"web-v\"\n"), 0o644)
	c := start(t, &Server{Version: "1", Dir: root})
	got := c.call("config.get", map[string]any{"setting": "projects.web.tag_prefix"})["result"].(map[string]any)
	if got["set"] != true || got["value"] != `"web-v"` || got["type"] != "string" {
		t.Fatalf("config.get answered %v", got)
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
	attest := c.call("config.get", map[string]any{"setting": "ci.attest"})["result"].(map[string]any)
	options, _ := attest["options"].([]any)
	if len(options) != 3 {
		t.Fatalf("ci.attest takes three values, each explained: %v", attest)
	}
	first := options[0].(map[string]any)
	if first["value"] != "on" || first["default"] != true || first["description"] == "" {
		t.Fatalf("the first option of ci.attest is on, explained and the default: %v", first)
	}
	if reply := c.call("config.get", map[string]any{"setting": "stacks.web"}); reply["error"].(map[string]any)["code"].(float64) != codeInvalidParams {
		t.Fatalf("an unknown setting answered %v", reply)
	}
}
