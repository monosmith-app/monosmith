package plugindraft

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestInspectFindsManifestTasksAndPublish(t *testing.T) {
	sample := fstest.MapFS{
		"zig.json":                {Data: []byte("{\n  \"name\": \"hello\",\n  \"version\": \"0.1.0\",\n  \"scripts\": {\"test\": \"zig test\", \"other\": \"x\"}\n}")},
		"Makefile":                {Data: []byte("build:\n\tzig build\nVAR := 1\ntest: build\n\tzig test\n.PHONY: build\n")},
		"README.md":               {Data: []byte("name: nope\nversion: 1")},
		"notes.cfg":               {Data: []byte("nothing here")},
		".github/workflows/p.yml": {Data: []byte("jobs:\n  x:\n    steps:\n      - run: zig publish --token $T\n")},
	}
	d := Inspect(sample)
	if len(d.Manifests) != 1 || d.Manifests[0].File != "zig.json" || !strings.Contains(d.Manifests[0].VersionLine, "0.1.0") {
		t.Errorf("manifests: %+v", d.Manifests)
	}
	got := map[string]string{}
	for _, task := range d.Tasks {
		got[task.Name+"@"+task.Source] = strings.Join(task.Command, " ")
	}
	for k, want := range map[string]string{"build@Makefile": "make build", "test@Makefile": "make test", "test@zig.json scripts": "zig test"} {
		if got[k] != want {
			t.Errorf("%s: got %q, want %q (all: %v)", k, got[k], want, got)
		}
	}
	if len(got) != 3 {
		t.Errorf("an unknown script or target was taken: %v", got)
	}
	if len(d.Publish) != 1 || !strings.Contains(d.Publish[0], "p.yml") {
		t.Errorf("publish: %v", d.Publish)
	}
}

func TestRenderIsTomlAndTodoNamesMissingParts(t *testing.T) {
	text := Render("zig", "sample", Draft{Manifests: []Manifest{{File: "zig.json", NameLine: `"name": "x"`, VersionLine: `"version": "1"`}}, Tasks: []Task{{Name: "build", Command: []string{"zig", "build"}, Source: "Makefile"}}})
	for _, want := range []string{`name = "lang-zig"`, `manifests = ["zig.json"]`, "[tasks.build] # from Makefile", `command = ["zig", "build"]`} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in\n%s", want, text)
		}
	}
	todo := strings.Join(Todo(Draft{}), "\n")
	if !strings.Contains(todo, "manifests: none was found") || !strings.Contains(todo, "no publish step") {
		t.Errorf("todo: %s", todo)
	}
}
