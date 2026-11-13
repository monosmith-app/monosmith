package tui

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

// projectActionsLab is a Projects tab with a Tauri app and an app that ships
// as an image beside the sample projects, and the forms smith builds from the
// descriptions of `smith kind`, `smith eject` and `smith tauri add`.
func projectActionsLab(t *testing.T) *lab {
	t.Helper()
	l := newLab(t, true, 140, 60)
	sh := l.sh()
	projects := append(slices.Clone(sampleProjects),
		projectlist.Project{Name: "desk", Kind: "app", Language: "node", Root: "apps/desk", Stack: "react-tauri",
			Tasks: []projectlist.Task{{Name: "build"}, {Name: "tauri-build"}}},
		projectlist.Project{Name: "api", Kind: "app", Language: "go", Root: "apps/api",
			Tasks: []projectlist.Task{{Name: "build"}, {Name: "image"}}},
	)
	for _, f := range []struct {
		desc clicontract.CommandDescription
		path []string
		id   string
	}{
		{clicontract.CommandDescription{Name: "kind", Description: "Make a project an app, an internal lib or a published package", Arguments: []clicontract.ArgumentDescription{{Name: "project", Required: true}, {Name: "kind", Required: true, Choices: []string{"app", "lib", "package"}}}}, []string{"kind"}, "kind"},
		{clicontract.CommandDescription{Name: "eject", Description: "Copy smith's default configuration of a tool into the repository", Arguments: []clicontract.ArgumentDescription{{Name: "tool", Variadic: true}}}, []string{"eject"}, "eject"},
		{clicontract.CommandDescription{Name: "add", Description: "Make an existing Node app a Tauri desktop app", Arguments: []clicontract.ArgumentDescription{{Name: "project", Required: true}}}, []string{"tauri", "add"}, "tauri add"},
	} {
		sh.guide = sh.guide.WithAction(guide.FormFor(f.desc, f.path, f.id))
	}
	l.m = sh
	return l.send(sources.ProjectsLoaded{Projects: projects, Branch: "main"})
}

// pick moves the selection of the list to the project with the name.
func (l *lab) pick(name string) *lab {
	l.t.Helper()
	for range 20 {
		l.keys("up")
	}
	for range 20 {
		if l.sh().selectedProject() == name {
			return l
		}
		l.keys("down")
	}
	l.t.Fatalf("no project %q in the list", name)
	return l
}

// actionIDs are the ids of the buttons of the detail, in their order.
func (l *lab) actionIDs() []string {
	var ids []string
	for _, a := range l.sh().detailActions() {
		ids = append(ids, a.ID)
	}
	return ids
}

// activateAction runs a button of the detail as a click on it would.
func (l *lab) activateAction(id string) *lab {
	sh := l.sh()
	cmd := sh.activate(id)
	l.m = sh
	l.drive(cmd)
	return l
}

func TestTheDetailOfAProjectShowsOnlyTheFormsThatApplyToIt(t *testing.T) {
	cases := []struct {
		project   string
		want, not []string
	}{
		// A Node web app with a stack: swap, environments, Add Tauri.
		{"web", []string{"proj.swap", "ws.env-add", "ws.env-remove", "x:tauri-add"}, []string{"ws.tauri-icon", "ws.docker-description"}},
		// A Tauri app: its icons, never Add Tauri again.
		{"desk", []string{"proj.swap", "ws.tauri-icon", "ws.env-add"}, []string{"x:tauri-add", "ws.docker-description"}},
		// A Go app that ships as an image: the Docker Hub descriptions only.
		{"api", []string{"ws.docker-description"}, []string{"proj.swap", "ws.env-add", "ws.tauri-icon", "x:tauri-add"}},
		// A Go app without a stack: none of them.
		{"orders", nil, []string{"proj.swap", "ws.env-add", "ws.env-remove", "ws.tauri-icon", "x:tauri-add", "ws.docker-description"}},
	}
	for _, c := range cases {
		l := projectActionsLab(t).pick(c.project)
		ids := l.actionIDs()
		for _, id := range append(c.want, "x:explain", "x:kind", "x:eject", "proj.remove") {
			if !slices.Contains(ids, id) {
				t.Errorf("%s: no button %s in %v", c.project, id, ids)
			}
		}
		for _, id := range c.not {
			if slices.Contains(ids, id) {
				t.Errorf("%s: button %s does not apply, got %v", c.project, id, ids)
			}
		}
	}
}

func TestMoreHoldsChangeKindEjectAndRemoveUnderTheProjectActions(t *testing.T) {
	l := projectActionsLab(t).pick("web")
	l.has("More", "Change kind…", "Eject a tool's configuration…", "Remove project…", "How was this resolved?")
	ids := l.actionIDs()
	if got := ids[len(ids)-3:]; !slices.Equal(got, []string{"x:kind", "x:eject", "proj.remove"}) {
		t.Fatalf("More is the last group, got %v", got)
	}
}

func TestWithoutTheirFormsKindEjectAndAddTauriHaveNoButton(t *testing.T) {
	l := newLab(t, true, 140, 60).pick("web")
	ids := l.actionIDs()
	for _, id := range []string{"x:kind", "x:eject", "x:tauri-add"} {
		if slices.Contains(ids, id) {
			t.Fatalf("%s shows with no form behind it: %v", id, ids)
		}
	}
	if !slices.Contains(ids, "proj.remove") {
		t.Fatalf("Remove is always there: %v", ids)
	}
}

func TestTheRemoveButtonOpensTheRemoveFormOnTheSelectedProject(t *testing.T) {
	l := projectActionsLab(t).pick("web")
	l.runDetailAction("[Remove project…")
	if l.sh().tab != commands.TabGuide || l.sh().guide.Title() != "Remove a project" {
		t.Fatalf("the click opened %q on %s", l.sh().guide.Title(), l.sh().tab)
	}
	l.keys("enter")
	if got := l.sh().guide.Answers()["name"]; !slices.Equal(got, []string{"web"}) {
		t.Fatalf("the form starts on the selected project, got %v", got)
	}
}

func TestTheProjectFormButtonsOpenTheirFormOnTheSelectedProject(t *testing.T) {
	cases := []struct{ project, id, key string }{
		{"desk", "x:kind", "arg:project"},
		{"web", "ws.env-add", "project"},
		{"web", "ws.env-remove", "project"},
	}
	for _, c := range cases {
		l := projectActionsLab(t).pick(c.project).activateAction(c.id)
		if l.sh().tab != commands.TabGuide {
			t.Fatalf("%s opened no form", c.id)
		}
		l.keys("enter")
		if got := l.sh().guide.Answers()[c.key]; !slices.Equal(got, []string{c.project}) {
			t.Fatalf("%s: the form starts on %s, got %v", c.id, c.project, got)
		}
	}
}

func TestTheAddTauriButtonAsksAboutTheSelectedNodeApp(t *testing.T) {
	l := projectActionsLab(t).pick("web").activateAction("x:tauri-add")
	if l.sh().guide.Title() != "Make an existing Node app a Tauri desktop app" {
		t.Fatalf("opened %q", l.sh().guide.Title())
	}
	l.keys("enter")
	l.has("smith tauri add web", "[Cancel (Esc)]")
	if len(l.started) != 0 {
		t.Fatalf("the confirmation comes before any change: %v", l.started)
	}
}

func TestTheTauriIconButtonOpensItsFormWithTheTauriAppPicked(t *testing.T) {
	l := projectActionsLab(t).pick("desk").activateAction("ws.tauri-icon")
	if l.sh().guide.Title() != "Make the icons of a Tauri app" {
		t.Fatalf("opened %q", l.sh().guide.Title())
	}
	l.typeText("logo.png").keys("enter")
	l.has("desk")
	l.keys("enter")
	l.has("smith tauri icon logo.png -p desk --yes", "[Cancel (Esc)]")
	if len(l.started) != 0 {
		t.Fatalf("the confirmation comes before any change: %v", l.started)
	}
}

func TestTheEjectAndDockerButtonsOpenTheirForms(t *testing.T) {
	l := projectActionsLab(t).pick("web").activateAction("x:eject")
	if l.sh().tab != commands.TabGuide || l.sh().guide.Title() != "Copy smith's default configuration of a tool into the repository" {
		t.Fatalf("eject opened %q", l.sh().guide.Title())
	}
	l = projectActionsLab(t).pick("api").activateAction("ws.docker-description")
	if l.sh().guide.Title() != "Set the Docker Hub descriptions" {
		t.Fatalf("docker opened %q", l.sh().guide.Title())
	}
}
