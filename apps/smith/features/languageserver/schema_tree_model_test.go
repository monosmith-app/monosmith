package languageserver

import "testing"

func TestLoadSchemaResolvesTablesKeysAndRefs(t *testing.T) {
	root, err := LoadSchema()
	if err != nil {
		t.Fatal(err)
	}
	if n := root.Resolve([]string{"workspace", "name"}); n == nil || n.Type != "string" || n.Description == "" {
		t.Errorf("workspace.name: %+v", n)
	}
	if n := root.Resolve([]string{"ci", "providers"}); n == nil || len(n.Items.Enum) == 0 {
		t.Errorf("ci.providers items enum: %+v", n)
	}
	// A project name is free; its keys come from additionalProperties.
	if n := root.Resolve([]string{"projects", "web", "tasks", "build", "command"}); n == nil {
		t.Errorf("task fields behind a $ref were not resolved")
	}
	if root.Resolve([]string{"workspace", "nope"}) != nil {
		t.Error("an unknown key resolved")
	}
	if !root.Resolve([]string{"ci"}).IsTable() || root.Resolve([]string{"ci", "devcontainer"}).IsTable() {
		t.Error("IsTable is wrong")
	}
}
