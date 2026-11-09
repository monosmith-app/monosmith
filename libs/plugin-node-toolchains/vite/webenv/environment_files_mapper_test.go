package webenv

import (
	"strings"
	"testing"
)

func TestSettingsDefaults(t *testing.T) {
	n := Settings{Names: DefaultNames}.Normalized()
	if n.Default != "prd" || n.Output != OutputWorkspace || n.EnvFiles != FilesCommit {
		t.Errorf("got %+v", n)
	}
	if got := (Settings{Names: []string{"dev", "stage"}}).Normalized().Default; got != "stage" {
		t.Errorf("without prd the last environment is the default, got %q", got)
	}
}

func TestSettingsRoundTripThroughTheTable(t *testing.T) {
	want := Settings{Names: DefaultNames, Default: "uat", Output: OutputProject, EnvFiles: FilesExample}
	got, ok := FromMap(want.ToMap())
	if !ok || strings.Join(got.Names, ",") != "lcl,dev,uat,prd" || got.Default != "uat" || got.Output != OutputProject || got.EnvFiles != FilesExample {
		t.Errorf("got %+v %v", got, ok)
	}
	if _, ok := FromMap(map[string]any{}); ok {
		t.Error("a table without environments is a single build")
	}
}

func TestValidateNames(t *testing.T) {
	for _, tc := range []struct {
		names []string
		bad   string
	}{
		{DefaultNames, ""},
		{[]string{"qa", "stage-2"}, ""},
		{nil, "at least one"},
		{[]string{"UAT"}, "not an environment name"},
		{[]string{"1st"}, "not an environment name"},
		{[]string{"dev", "dev"}, "twice"},
		{[]string{"production"}, "already give it a meaning"},
		{[]string{"local"}, "already give it a meaning"},
	} {
		err := ValidateNames(tc.names)
		if tc.bad == "" && err != nil || tc.bad != "" && (err == nil || !strings.Contains(err.Error(), tc.bad)) {
			t.Errorf("%v: got %v, want %q", tc.names, err, tc.bad)
		}
	}
	if err := (Settings{Names: DefaultNames, Default: "qa"}).Validate(); err == nil || !strings.Contains(err.Error(), "pick one of them") {
		t.Errorf("a default outside the list: %v", err)
	}
	if got := ParseList("lcl, dev uat;prd,"); strings.Join(got, "|") != "lcl|dev|uat|prd" {
		t.Errorf("got %v", got)
	}
}

func TestScriptsBuildEachEnvironmentIntoItsOwnFolder(t *testing.T) {
	got := Scripts(Settings{Names: DefaultNames}, "apps/web")
	want := map[string]string{
		"build":     "vite build --mode prd --outDir ../../dist/apps/web/prd",
		"build:lcl": "vite build --mode lcl --outDir ../../dist/apps/web/lcl",
		"build:dev": "vite build --mode dev --outDir ../../dist/apps/web/dev",
		"build:uat": "vite build --mode uat --outDir ../../dist/apps/web/uat",
		"start":     "vite --mode lcl",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
	inside := Scripts(Settings{Names: []string{"dev", "prd"}, Output: OutputProject}, "apps/web")
	if inside["build:dev"] != "vite build --mode dev --outDir dist/dev" || inside["build"] != "vite build --mode prd --outDir dist/prd" {
		t.Errorf("got %v", inside)
	}
	if _, ok := inside["start"]; ok {
		t.Error("start keeps the plain dev server when there is no lcl")
	}
}

func TestFilesWriteAnEnvFilePerEnvironment(t *testing.T) {
	files := Files(Settings{Names: DefaultNames})
	for _, name := range []string{".env.lcl", ".env.dev", ".env.uat", ".env.prd", "vite.build.ts", ".gitignore"} {
		if files[name] == "" {
			t.Errorf("no %s", name)
		}
	}
	if _, ok := files[".env.example"]; ok {
		t.Error("an example file only when the env files are not committed")
	}
	if !strings.Contains(files[".env.uat"], "never put a secret here") || !strings.Contains(files[".env.uat"], "VITE_API_URL=") {
		t.Errorf("the starter warns and has a placeholder:\n%s", files[".env.uat"])
	}
	if !strings.Contains(files[".gitignore"], ".env.*.local") || strings.Contains(files[".gitignore"], "!.env.example") {
		t.Errorf("gitignore:\n%s", files[".gitignore"])
	}
	example := Files(Settings{Names: DefaultNames, EnvFiles: FilesExample})
	if example[".env.example"] == "" || !strings.Contains(example[".gitignore"], "!.env.example") {
		t.Errorf("only the example is committed: %v", example[".gitignore"])
	}
}

func TestBuildOptionsMatchTheEnvironments(t *testing.T) {
	text := BuildOptionsFile(Settings{Names: DefaultNames})
	for _, want := range []string{"export function build (mode: string)", "case 'lcl'", "options.sourcemap = true", "case 'dev'", "minify:      true", "terser"} {
		if !strings.Contains(text, want) {
			t.Errorf("vite.build.ts misses %q:\n%s", want, text)
		}
	}
	if strings.Contains(BuildOptionsFile(Settings{Names: []string{"uat", "prd"}}), "case 'lcl'") {
		t.Error("no lcl, no lcl case")
	}
}
