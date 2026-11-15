package guide

import (
	"reflect"
	"testing"
)

func TestInstallPackageArgv(t *testing.T) {
	tests := []struct {
		name    string
		answers map[string][]string
		want    []string
	}{
		{"one project, prod", map[string][]string{"names": {"zod react@19"}, "projects": {"web"}, "kind": {"prod"}}, []string{"install", "-p", "web", "zod", "react@19"}},
		{"two projects, dev", map[string][]string{"names": {"vitest"}, "projects": {"web", "api"}, "kind": {"dev"}}, []string{"install", "-p", "web", "-p", "api", "-D", "vitest"}},
		{"root and a project, dev", map[string][]string{"names": {"vitest"}, "projects": {".", "web"}, "kind": {"dev"}}, []string{"install", "-p", ".", "-p", "web", "-D", "vitest"}},
		{"no project", map[string][]string{"names": {"zod"}, "kind": {"prod"}}, nil},
		{"no name", map[string][]string{"projects": {"web"}}, nil},
	}
	for _, tt := range tests {
		if got := InstallPackageArgv(tt.answers); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestInstallNamesRefuseAnOption(t *testing.T) {
	if installNamesProblem(nil, "zod -D") == "" {
		t.Fatal("an option typed as a name must be refused")
	}
	if installNamesProblem(nil, "zod @scope/x@1") != "" {
		t.Fatal("names with a version are fine")
	}
}

func TestInstallFormOffersTheRootFirst(t *testing.T) {
	a := InstallPackageAction([]InstallTarget{RootInstallTarget, {Label: "web", Value: "web"}}, []string{"."})
	var projects Step
	for _, st := range a.Steps {
		if st.Key == "projects" {
			projects = st
		}
	}
	if len(projects.Options) != 2 || projects.Options[0].Value != "." {
		t.Fatalf("the root is the first option: %v", projects.Options)
	}
	if got := projects.Ticked(nil); len(got) != 1 || got[0] != "." {
		t.Fatalf("a root row presets the root: %v", got)
	}
}
