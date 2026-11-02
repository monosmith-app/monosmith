package toolinstall

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolcatalog"
)

// onPath is a fake PATH: only the named programs exist.
func onPath(programs ...string) LookPath {
	return func(p string) (string, error) {
		for _, have := range programs {
			if have == p {
				return p, nil
			}
		}
		return "", errors.New("not found")
	}
}

func missing(name string) Status {
	t, _ := toolcatalog.Find(name)
	return Status{Need: Need{ToolNeed: t}, State: Missing}
}

func optionValues(p ToolPlan) []string {
	var out []string
	for _, o := range p.Options {
		out = append(out, o.Value)
	}
	return out
}

func TestDetectManagersPerSystem(t *testing.T) {
	all := []string{"winget", "brew", "apt-get", "dnf", "pacman", "uv", "fnm", "rustup", "npm"}
	names := func(ms []Manager) []string {
		out := []string{}
		for _, m := range ms {
			out = append(out, m.Name)
		}
		return out
	}
	for goos, want := range map[string][]string{
		"windows": {"winget", "uv", "fnm", "rustup", "npm"},
		"darwin":  {"brew", "brew-cask", "uv", "fnm", "rustup", "npm"},
		"linux":   {"brew", "apt", "dnf", "pacman", "uv", "fnm", "rustup", "npm"},
	} {
		if got := names(DetectManagers(goos, onPath(all...))); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %v, want %v", goos, got, want)
		}
	}
	if got := DetectManagers("windows", onPath()); len(got) != 0 {
		t.Errorf("no program on PATH: got %v", got)
	}
}

func TestBuildCommandPerManager(t *testing.T) {
	user := Env{OS: "linux"}
	for name, tc := range map[string]struct {
		manager, pkg string
		env          Env
		want         string
		admin        bool
	}{
		"winget":         {"winget", "Python.Python.3.13", Env{OS: "windows"}, "winget install --id Python.Python.3.13 -e", true},
		"winget script":  {"winget", "Git.Git", Env{OS: "windows", NonInteractive: true}, "winget install --id Git.Git -e --accept-package-agreements --accept-source-agreements --disable-interactivity", true},
		"brew":           {"brew", "go", Env{OS: "darwin"}, "brew install go", false},
		"brew cask":      {"brew-cask", "dotnet-sdk", Env{OS: "darwin"}, "brew install --cask dotnet-sdk", false},
		"apt needs sudo": {"apt", "nodejs npm", user, "sudo apt-get install nodejs npm", true},
		"apt as root":    {"apt", "git", Env{OS: "linux", Root: true}, "apt-get install git", false},
		"dnf script":     {"dnf", "git", Env{OS: "linux", NonInteractive: true}, "sudo dnf install -y git", true},
		"pacman script":  {"pacman", "go", Env{OS: "linux", NonInteractive: true}, "sudo pacman -S --noconfirm go", true},
		"uv":             {"uv", "3.13", user, "uv python install 3.13", false},
		"fnm":            {"fnm", "22", user, "fnm install 22", false},
		"rustup":         {"rustup", "stable", user, "rustup default stable", false},
		"npm":            {"npm", "azure-functions-core-tools@4", user, "npm install -g azure-functions-core-tools@4", false},
	} {
		got, ok := BuildCommand(tc.manager, tc.pkg, tc.env)
		if !ok || got.String() != tc.want || got.Admin != tc.admin {
			t.Errorf("%s: got %q admin=%v ok=%v, want %q admin=%v", name, got.String(), got.Admin, ok, tc.want, tc.admin)
		}
	}
	if _, ok := BuildCommand("chocolatey", "x", user); ok {
		t.Error("an unknown manager must not build a command")
	}
}

func TestBuildPlanPerSystem(t *testing.T) {
	all := []string{"winget", "brew", "apt-get", "dnf", "pacman", "uv", "fnm", "rustup", "npm"}
	for name, tc := range map[string]struct {
		tool, goos  string
		have        []string
		wantOptions []string
		wantDefault string
	}{
		"python on windows, winget only":   {"python", "windows", []string{"winget"}, []string{"install:winget", "page", "myself", "skip"}, "install:winget"},
		"python on windows with uv":        {"python", "windows", []string{"winget", "uv"}, []string{"install:uv", "install:winget", "page", "myself", "skip"}, "install:uv"},
		"python on mac with brew":          {"python", "darwin", []string{"brew"}, []string{"install:brew", "page", "myself", "skip"}, "install:brew"},
		"python on debian":                 {"python", "linux", []string{"apt-get"}, []string{"install:apt", "page", "myself", "skip"}, "install:apt"},
		"python on a linux with brew+apt":  {"python", "linux", []string{"apt-get", "brew"}, []string{"install:brew", "install:apt", "page", "myself", "skip"}, "install:brew"},
		"node prefers fnm over winget":     {"node", "windows", []string{"winget", "fnm"}, []string{"install:fnm", "install:winget", "page", "myself", "skip"}, "install:fnm"},
		"go on arch":                       {"go", "linux", []string{"pacman"}, []string{"install:pacman", "page", "myself", "skip"}, "install:pacman"},
		"no package manager":               {"go", "linux", nil, []string{"page", "myself", "skip"}, "myself"},
		"docker on mac is a cask":          {"docker", "darwin", []string{"brew"}, []string{"install:brew-cask", "page", "myself", "skip"}, "install:brew-cask"},
		"dart on linux has no manager way": {"dart", "linux", []string{"apt-get", "brew"}, []string{"page", "myself", "skip"}, "myself"},
	} {
		plan := BuildPlan(missing(tc.tool), DetectManagers(tc.goos, onPath(append(tc.have, all[:0]...)...)), Env{OS: tc.goos})
		if got := optionValues(plan); !reflect.DeepEqual(got, tc.wantOptions) {
			t.Errorf("%s: options %v, want %v", name, got, tc.wantOptions)
		}
		if plan.Default != tc.wantDefault {
			t.Errorf("%s: default %q, want %q", name, plan.Default, tc.wantDefault)
		}
	}
}

func TestBuildPlanPrefersANoAdminOptionAsTheDefault(t *testing.T) {
	// On linux with apt (sudo) and brew (no admin) the default is brew even though apt is listed first for some tools.
	plan := BuildPlan(missing("git"), DetectManagers("linux", onPath("apt-get", "brew")), Env{OS: "linux"})
	if plan.Default != "install:brew" {
		t.Fatalf("default %q", plan.Default)
	}
	apt, _ := plan.Find("install:apt")
	if !apt.Admin || !strings.Contains(apt.Hint, "administrator") {
		t.Fatalf("apt must say it needs administrator rights: %+v", apt)
	}
}

func TestBuildPlanFollowsThePinnedVersion(t *testing.T) {
	st := missing("python")
	st.Wanted = Requirement{Min: "3.11"}
	plan := BuildPlan(st, DetectManagers("windows", onPath("winget")), Env{OS: "windows"})
	o, _ := plan.Find("install:winget")
	if got := strings.Join(o.Command, " "); got != "winget install --id Python.Python.3.11 -e" {
		t.Fatalf("command %q", got)
	}
}

func TestEveryCatalogueEntryBuildsACommandOnItsSystems(t *testing.T) {
	for _, tool := range toolcatalog.All() {
		for _, in := range tool.Install {
			if _, ok := BuildCommand(in.Manager, Fill(in.Package, tool.Suggested), Env{OS: "linux"}); !ok {
				t.Errorf("%s: no command for manager %s", tool.Name, in.Manager)
			}
			if strings.Contains(Fill(in.Package, tool.Suggested), "{") {
				t.Errorf("%s: placeholder left in %q", tool.Name, in.Package)
			}
		}
		if err := protocol.ValidateToolNeed(tool); err != nil {
			t.Error(err)
		}
	}
}
