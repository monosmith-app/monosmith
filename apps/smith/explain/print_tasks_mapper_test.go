package explain

import (
	"bytes"
	"strings"
	"testing"
)

func sampleTasks() []Task {
	field := func(name, value, source string) Field { return Field{name, value, source} }
	task := func(id, command, deps, cache string) Task {
		return Task{ID: id, Fields: []Field{
			field("command", command, "plugin node"),
			field("cwd", "packages/env", "plugin node"),
			field("dependsOn", deps, "plugin node"),
			field("inputs", "**/* (every file of the project)", "plugin node"),
			field("outputs", "—", "plugin node"),
			field("env", "—", "plugin node"),
			field("cache", cache, "plugin node"),
		}}
	}
	tasks := []Task{task("env:build", "npm run build", "^build", "true"), task("env:deps", "{smith} npm-deps-check", "—", "false")}
	tasks[1].Fields[6].Source = "smith.toml [tasks.deps]"
	return tasks
}

func TestPrintTasksStartsWithAOneLineSummaryOfEachTask(t *testing.T) {
	var out bytes.Buffer
	printTasks(&out, sampleTasks())
	text := out.String()
	if !strings.HasPrefix(text, "env has 2 tasks:\n  build  npm run build  (cached, after ^build)\n  deps   smith npm-deps-check\n") {
		t.Fatalf("%s", text)
	}
}

func TestPrintTasksSaysWhereValuesCameFromOnce(t *testing.T) {
	var out bytes.Buffer
	printTasks(&out, sampleTasks())
	text := out.String()
	if n := strings.Count(text, "plugin node"); n != 1 {
		t.Errorf("provenance repeated %d times:\n%s", n, text)
	}
	if !strings.Contains(text, "(from smith.toml [tasks.deps])") || !strings.Contains(text, "Every value above comes from plugin node, except where marked.") {
		t.Errorf("%s", text)
	}
}

func TestPrintTasksExplainsWhatANewcomerCannotRead(t *testing.T) {
	var out bytes.Buffer
	printTasks(&out, sampleTasks())
	text := out.String()
	if strings.Contains(text, "{smith}") || !strings.Contains(text, "^build (build of the projects it imports, first)") {
		t.Fatalf("%s", text)
	}
}
