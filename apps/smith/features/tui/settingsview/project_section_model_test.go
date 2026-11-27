package settingsview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

var webRef = ProjectRef{Name: "web", Root: "apps/web", Language: "node", Kind: "app", Version: "1.2.0"}

// projectSample is the sample workspace whose backend also reads the
// settings of web: two read-only values from its manifest and two settings.
func projectSample() (*fake, Backend) {
	f := sample()
	b := f.backend()
	tagPrefix := Setting{Path: "projects.web.tag_prefix", Name: "tag_prefix", Type: "string", Explanation: "Replaces the release tag prefix."}
	b.LoadProject = func(p ProjectRef) ProjectView {
		f.calls = append(f.calls, "load "+p.Name)
		st := tagPrefix
		for _, c := range f.calls {
			if strings.HasPrefix(c, "set projects.web.tag_prefix ") {
				st.Value, st.Set, st.Current = `"web-v"`, true, []string{"web-v"}
			}
		}
		return ProjectView{
			Name: p.Name, Manifest: "apps/web/package.json",
			Facts: []Fact{
				{Key: "name", Value: "web", Explanation: "The project's name, as its manifest gives it.", Source: "apps/web/package.json"},
				{Key: "version", Value: "1.2.0", Explanation: "The version its manifest records.", Source: "apps/web/package.json"},
			},
			Settings: []Setting{
				{Path: "projects.web.bare_tags", Name: "bare_tags", Type: "boolean", Explanation: "Tag releases with the bare version."},
				st,
			},
		}
	}
	return f, b
}

func openedProject(t *testing.T) (*fake, Model) {
	t.Helper()
	f, b := projectSample()
	m, cmd := New(b, nil).Open()
	m, _ = settle(m, Result{Cmd: cmd})
	m, cmd = m.OpenProject(webRef)
	m, _ = settle(m, Result{Cmd: cmd})
	return f, m
}

func TestTheProjectSectionShowsTheManifestValuesReadOnlyAndTheProjectsSettings(t *testing.T) {
	_, m := openedProject(t)
	if m.Section() != Project || m.ProjectOpen() != "web" {
		t.Fatalf("section %v, project %q", m.Section(), m.ProjectOpen())
	}
	out := text(m, 110, 30)
	for _, want := range []string{"Project web", "From apps/web/package.json (read-only)", "version", "1.2.0", "[projects.web] in smith.toml", "bare_tags", "tag_prefix"} {
		if !strings.Contains(out, want) {
			t.Fatalf("the section lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "projects.web.tag_prefix  ") {
		t.Fatalf("a row names the setting, not its whole path:\n%s", out)
	}
}

func TestEnterOnAReadOnlyValueSaysWhereToChangeItAndWritesNothing(t *testing.T) {
	f, m := openedProject(t)
	m, _ = press(m, "down", "enter") // version
	if !strings.Contains(m.note, "version comes from apps/web/package.json") {
		t.Fatalf("note %q", m.note)
	}
	for _, c := range f.calls {
		if strings.HasPrefix(c, "set ") || strings.HasPrefix(c, "unset ") {
			t.Fatalf("a read-only value wrote %q", c)
		}
	}
	if m.Reasons()["settings.change"] == "" {
		t.Fatal("the footer must say a read-only value cannot be changed")
	}
}

func TestAProjectSettingIsWrittenAtItsPathAndTheSectionReadsItAgain(t *testing.T) {
	f, m := openedProject(t)
	m, _ = press(m, "end", "enter")
	for _, r := range "web-v" {
		m, _ = press(m, string(r))
	}
	m, _ = press(m, "enter")
	if got := f.calls[len(f.calls)-2]; got != "set projects.web.tag_prefix web-v" {
		t.Fatalf("calls %v", f.calls)
	}
	if f.calls[len(f.calls)-1] != "load web" {
		t.Fatalf("the section must read the project again after the change: %v", f.calls)
	}
	if out := text(m, 110, 30); !strings.Contains(out, `"web-v"`) {
		t.Fatalf("the new value is not shown:\n%s", out)
	}
	_, _ = press(m, "u")
	if got := f.calls[len(f.calls)-2]; got != "unset projects.web.tag_prefix" {
		t.Fatalf("u must clear the key: %v", f.calls)
	}
}

func TestAProjectSettingTheValidatorsRefuseShowsTheirMessageInTheEditor(t *testing.T) {
	f, m := openedProject(t)
	f.setErr = errTest("smith.toml: web and api all set bare_tags, but only one project can tag releases with the bare version")
	m, _ = press(m, "down", "down", "enter") // bare_tags: the switch is a choice
	if !m.Modal() {
		t.Fatal("Enter on a switch presents its choice")
	}
	m, _ = press(m, "down", "enter") // from Not decided to On, saved
	if out := pane.Strip(strings.Join(m.ProjectPane(plainTheme(), 100, 30, -1, -1), "\n")); !m.Modal() || !strings.Contains(out, "only one project") {
		t.Fatalf("the refusal shows in the open choice:\n%s", out)
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }
