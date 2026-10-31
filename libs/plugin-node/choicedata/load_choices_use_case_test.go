package choicedata

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const catalog = `
[versions]
react = "^19.0.0"
vite = "^8.0.0"
bootstrap = "^5.3.0"
`

const base = `
dev_dependencies = ["vite"]
[tasks]
build = "vite build"
[slots]
"vite.plugins" = []
[files]
"index.html" = '''<div id="app"></div>'''
"src/main.ts" = '''import './base'
'''
`

const reactToml = `
axis = "framework"
name = "react"
description = "React"
language = "node"
roles = ["scaffold", "toolchain"]
needs = ["test"]
dependencies = ["react"]
[slots]
"vite.plugins" = ["react()"]
[files]
"src/main.ts" = '''import './react'
'''
`

const bootstrapToml = `
axis = "style"
name = "bootstrap"
language = "node"
dependencies = ["bootstrap"]
[slots]
"entry.imports" = ["import 'bootstrap/dist/css/bootstrap.min.css'"]
`

func fixture(extra map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{
		"d/catalog.toml":   {Data: []byte(catalog)},
		"d/base.toml":      {Data: []byte(base)},
		"d/react.toml":     {Data: []byte(reactToml)},
		"d/bootstrap.toml": {Data: []byte(bootstrapToml)},
		"d/.gitkeep":       {Data: nil},
	}
	for name, body := range extra {
		fsys["d/"+name] = &fstest.MapFile{Data: []byte(body)}
	}
	return fsys
}

func TestLoadPublishesTheChoicesAndTheStacks(t *testing.T) {
	s, err := Load(fixture(map[string]string{"stacks.toml": `
[[stack]]
name = "react"
description = "React app"
choices = ["react", "bootstrap"]
toolchain = ["tauri"]
`}), "d")
	if err != nil {
		t.Fatal(err)
	}
	want := []protocol.Choice{
		{Axis: "framework", Name: "react", Description: "React", Language: "node", Roles: []protocol.Role{protocol.RoleScaffold, protocol.RoleToolchain}, Needs: []string{"test"}},
		{Axis: "style", Name: "bootstrap", Language: "node"},
	}
	if got := s.Choices(); !reflect.DeepEqual(got, want) {
		t.Errorf("choices %+v, want %+v", got, want)
	}
	if got := s.Stacks(); len(got) != 1 || got[0].Name != "react" || got[0].Toolchain[0] != "tauri" || len(got[0].Choices) != 2 {
		t.Errorf("stacks %+v", got)
	}
}

func TestLoadAnEmptyDirectoryIsAnEmptySet(t *testing.T) {
	s, err := Load(fstest.MapFS{"d/.gitkeep": {}}, "d")
	if err != nil || len(s.Choices()) != 0 {
		t.Fatalf("%v %v", s, err)
	}
}

func TestLoadRefusesBadData(t *testing.T) {
	cases := map[string]struct {
		files map[string]string
		want  string
	}{
		"unknown member":  {map[string]string{"vue.toml": "axis = \"framework\"\nname = \"vue\"\nlanguage = \"node\"\nfacts = 1\n"}, "unknown member"},
		"name and file":   {map[string]string{"vue.toml": "axis = \"framework\"\nname = \"svelte\"\nlanguage = \"node\"\n"}, "must be named svelte.toml"},
		"no version":      {map[string]string{"vue.toml": "axis = \"framework\"\nname = \"vue\"\nlanguage = \"node\"\ndependencies = [\"vue\"]\n"}, "vue, which has no version in catalog.toml"},
		"bad role":        {map[string]string{"vue.toml": "axis = \"framework\"\nname = \"vue\"\nlanguage = \"node\"\nroles = [\"deploy\"]\n"}, "scaffold or toolchain"},
		"path leaves":     {map[string]string{"vue.toml": "axis = \"framework\"\nname = \"vue\"\nlanguage = \"node\"\n[files]\n\"../x\" = \"\"\n"}, "inside the project"},
		"bad toml":        {map[string]string{"vue.toml": "axis = "}, "vue.toml"},
		"empty stack":     {map[string]string{"stacks.toml": "[[stack]]\nname = \"a\"\n"}, "needs a name and choices"},
		"base dependency": {map[string]string{"base.toml": "dependencies = [\"left-pad\"]\n"}, "base names the package left-pad"},
	}
	for name, tc := range cases {
		_, err := Load(fixture(tc.files), "d")
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: got %v, want %q", name, err, tc.want)
		}
	}
}
