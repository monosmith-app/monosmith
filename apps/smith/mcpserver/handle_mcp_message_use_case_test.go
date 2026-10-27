package mcpserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func call(t *testing.T, s Server, msg string) map[string]any {
	t.Helper()
	raw := s.Handle([]byte(msg))
	if raw == nil {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestServer(t *testing.T) {
	var ran []string
	s := Server{Tools: Tools, Version: "1.2.3", Run: func(args []string) Outcome {
		ran = append(ran, strings.Join(args, " "))
		return Outcome{Text: "ok"}
	}}
	init := call(t, s, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}`)
	result := init["result"].(map[string]any)
	if result["protocolVersion"] != "2025-03-26" || result["serverInfo"].(map[string]any)["version"] != "1.2.3" {
		t.Fatalf("initialize %v", init)
	}
	if call(t, s, `{"jsonrpc":"2.0","method":"notifications/initialized"}`) != nil {
		t.Fatal("a notification has no answer")
	}
	listed := call(t, s, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)["result"].(map[string]any)["tools"].([]any)
	if len(listed) != len(Tools) {
		t.Fatalf("%d tools listed", len(listed))
	}
	for _, c := range []struct{ msg, want string }{
		{`{"name":"smith_projects"}`, "projects --json"},
		{`{"name":"smith_install","arguments":{"package":"zod@^3","projects":["web","apps/api"],"dev":true}}`, "install -w web -w apps/api -D zod@^3"},
		{`{"name":"smith_run","arguments":{"tasks":["test","lint"],"projects":["web"],"affected":true}}`, "run test lint -p web --affected"},
		{`{"name":"smith_add_project","arguments":{"kind":"app","name":"api","lang":"go"}}`, "add app api --lang go"},
		{`{"name":"smith_remove_project","arguments":{"name":"old","dry_run":true}}`, "remove old --dry-run"},
		{`{"name":"smith_change_kind_preview","arguments":{"project":"core","kind":"package"}}`, "kind core package --dry-run --json"},
		{`{"name":"smith_change_kind","arguments":{"project":"core","kind":"lib"}}`, "kind core lib --json"},
		{`{"name":"smith_pipeline","arguments":{"template":"deploy-pages","project":"web"}}`, "pipeline deploy-pages --project web"},
		{`{"name":"smith_graph","arguments":{"focus":["web","api"]}}`, "graph --format json --focus web,api"},
		{`{"name":"smith_explain","arguments":{"target":"web:build"}}`, "explain web:build --json"},
		{`{"name":"smith_explain","arguments":{"target":"web","recipe":true}}`, "explain web --json --recipe"},
		{`{"name":"smith_recipe_check","arguments":{"recipe":"team.recipe.toml","choices":["language=python","kind=library"],"all":true}}`, "recipe check team.recipe.toml --json --choice language=python --choice kind=library --all"},
		{`{"name":"smith_add_project","arguments":{"kind":"package","name":"billing","choices":["language=python"]}}`, "add package billing --choice language=python"},
		{`{"name":"smith_plugin_list"}`, "plugin list --json"},
		{`{"name":"smith_plugin_check","arguments":{"plugin":"kiwi","generate":true}}`, "plugin check kiwi --json --generate"},
		{`{"name":"smith_doctor_fixes"}`, "doctor --fix"},
		{`{"name":"smith_doctor_fix"}`, "doctor --fix --yes"},
		{`{"name":"smith_upgrade_preview"}`, "upgrade --dry-run --json"},
		{`{"name":"smith_upgrade"}`, "upgrade"},
		{`{"name":"smith_recipe_new","arguments":{"answers":"answers.toml","path":"recipes/team.recipe.toml","force":true}}`, "recipe new recipes/team.recipe.toml --from-answers answers.toml --force"},
	} {
		ran = nil
		res := call(t, s, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":`+c.msg+`}`)["result"].(map[string]any)
		if res["isError"] != false || len(ran) != 1 || ran[0] != c.want {
			t.Errorf("%s: ran %v, result %v", c.msg, ran, res)
		}
	}
	ran = nil
	res := call(t, s, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"smith_install","arguments":{"package":"--registry=evil","projects":["web"]}}}`)["result"].(map[string]any)
	if res["isError"] != true || len(ran) != 0 {
		t.Fatalf("an option smuggled in as a value must be refused: %v %v", res, ran)
	}
	if e := call(t, s, `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"nope"}}`)["error"]; e == nil {
		t.Fatal("unknown tool")
	}
	if e := call(t, s, `{"jsonrpc":"2.0","id":6,"method":"resources/list"}`)["error"].(map[string]any); e["code"] != float64(-32601) {
		t.Fatalf("unknown method %v", e)
	}
}

func TestTheNewToolsRefuseWhatCouldSmuggleInAnOptionOrRunAFile(t *testing.T) {
	var ran []string
	s := Server{Tools: Tools, Version: "1", Run: func(args []string) Outcome {
		ran = append(ran, strings.Join(args, " "))
		return Outcome{Text: "ok"}
	}}
	for _, msg := range []string{
		`{"name":"smith_plugin_check","arguments":{"plugin":"./bin/smith-plugin-kiwi"}}`,
		`{"name":"smith_plugin_check","arguments":{"plugin":"/usr/bin/evil"}}`,
		`{"name":"smith_plugin_check","arguments":{"plugin":"Kiwi"}}`,
		`{"name":"smith_recipe_check","arguments":{"recipe":"--json"}}`,
		`{"name":"smith_recipe_check","arguments":{"recipe":"r.toml","choices":["language"]}}`,
		`{"name":"smith_recipe_check","arguments":{"recipe":"r.toml","choices":["--x=y"]}}`,
		`{"name":"smith_add_project","arguments":{"kind":"app","name":"web","choices":["language="]}}`,
		`{"name":"smith_explain","arguments":{"target":"-p"}}`,
		`{"name":"smith_recipe_new","arguments":{"answers":"--force"}}`,
	} {
		res := call(t, s, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":`+msg+`}`)["result"].(map[string]any)
		if res["isError"] != true {
			t.Errorf("%s must be refused: %v", msg, res)
		}
	}
	if len(ran) != 0 {
		t.Errorf("nothing ran: %v", ran)
	}
}

func TestOnlyTheToolsThatChangeNothingAreMarkedReadOnly(t *testing.T) {
	readOnly := map[string]bool{}
	for _, tool := range Tools {
		readOnly[tool.Name] = tool.ReadOnly
	}
	for _, name := range []string{"smith_explain", "smith_recipe_check", "smith_plugin_list", "smith_doctor_fixes", "smith_upgrade_preview"} {
		if !readOnly[name] {
			t.Errorf("%s changes nothing", name)
		}
	}
	for _, name := range []string{"smith_doctor_fix", "smith_upgrade", "smith_recipe_new", "smith_plugin_check", "smith_add_project"} {
		if readOnly[name] {
			t.Errorf("%s changes things or runs a plugin, so a client should ask", name)
		}
	}
}
