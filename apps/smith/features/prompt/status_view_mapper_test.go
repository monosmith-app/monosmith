package prompt

import (
	"strings"
	"testing"
	"time"
)

func summaryOf(failed, warnings int, projects ...Project) Summary {
	return Summary{Schema: schemaVersion, Workspace: "/w", CheckedAt: time.Unix(0, 0), Failed: failed, Warnings: warnings, Projects: projects}
}

func TestLineWithoutASummaryIsTheVersionAloneAndSaysNothingWrong(t *testing.T) {
	got := NewView("1.2.3", Summary{}, false, []string{"api/main.go"}).Line()
	if got != "smith 1.2.3" {
		t.Fatalf("got %q", got)
	}
}

func TestLineShowsProjectsFailuresWarningsAndAffectedProjects(t *testing.T) {
	s := summaryOf(2, 3,
		Project{Name: "core", Root: "libs/core", Dependents: []string{"api", "web"}},
		Project{Name: "api", Root: "apps/api"},
		Project{Name: "web", Root: "apps/web"})
	got := NewView("1.2.3", s, true, []string{"libs/core/x.go"}).Line()
	want := "smith 1.2.3 · 3 projects · × 2 ▲ 3 · api, core, web"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestLineOmitsAZeroCountAndShowsTheTickWhenBothAreZero(t *testing.T) {
	one := summaryOf(0, 1, Project{Name: "api", Root: "api"})
	if got, want := NewView("1", one, true, nil).Line(), "smith 1 · 1 project · ▲ 1"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	clean := summaryOf(0, 0, Project{Name: "api", Root: "api"})
	if got, want := NewView("1", clean, true, nil).Line(), "smith 1 · 1 project · ✓"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestLineNamesThreeAffectedProjectsAndCountsTheRest(t *testing.T) {
	var projects []Project
	var files []string
	for _, n := range []string{"a", "b", "c", "d", "e"} {
		projects = append(projects, Project{Name: n, Root: n})
		files = append(files, n+"/f.go")
	}
	got := NewView("1", summaryOf(0, 0, projects...), true, files).Line()
	if !strings.HasSuffix(got, " · a, b, c +2") {
		t.Fatalf("got %q", got)
	}
}

func TestViewCarriesTheFullAffectedListForJSON(t *testing.T) {
	var projects []Project
	var files []string
	for _, n := range []string{"a", "b", "c", "d"} {
		projects = append(projects, Project{Name: n, Root: n})
		files = append(files, n+"/f.go")
	}
	v := NewView("1", summaryOf(1, 0, projects...), true, files)
	if len(v.Affected) != 4 || *v.Projects != 4 || v.Doctor.Failed != 1 {
		t.Fatalf("got %+v", v)
	}
}
