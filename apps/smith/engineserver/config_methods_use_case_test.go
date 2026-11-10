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
