package toolinstall

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func tool(name, choice string) protocol.ToolNeed {
	return protocol.ToolNeed{Name: name, Choice: choice}
}

func names(needs []Need) []string {
	out := []string{}
	for _, n := range needs {
		out = append(out, n.Name)
	}
	return out
}

func TestNeeded(t *testing.T) {
	core := []protocol.ToolNeed{tool("git", "")}
	py := Source{Plugin: "python", Languages: []string{"python", "py"}, Tools: []protocol.ToolNeed{tool("python", ""), tool("func", "azure-functions")}}
	node := Source{Plugin: "node", Languages: []string{"node"}, Tools: []protocol.ToolNeed{tool("node", ""), tool("func", "azure-functions")}}
	for name, tc := range map[string]struct {
		uses    []Use
		extraL  []string
		extraC  []string
		want    []string
		project []string // projects of the tool "python"
	}{
		"no projects: only the core": {nil, nil, nil, []string{"git"}, nil},
		"a python project":           {[]Use{{Project: "api", Language: "python", Root: "/w/api"}}, nil, nil, []string{"git", "python"}, []string{"api"}},
		"a language named by flag":   {nil, []string{"py"}, nil, []string{"git", "python"}, nil},
		"a choice needs its tool":    {[]Use{{Project: "fn", Language: "node", Choices: []string{"azure-functions"}}}, nil, nil, []string{"git", "node", "func"}, nil},
		"a choice named by flag":     {nil, nil, []string{"azure-functions"}, []string{"git", "func"}, nil},
		"two plugins, one tool once": {[]Use{{Project: "a", Language: "python", Choices: []string{"azure-functions"}}, {Project: "b", Language: "node", Choices: []string{"azure-functions"}}}, nil, nil, []string{"git", "python", "func", "node"}, []string{"a"}},
		"an unrelated language":      {[]Use{{Project: "x", Language: "go"}}, nil, nil, []string{"git"}, nil},
	} {
		got := Needed("/w", core, []Source{py, node}, tc.uses, tc.extraL, tc.extraC)
		if !reflect.DeepEqual(names(got), tc.want) {
			t.Errorf("%s: got %v, want %v", name, names(got), tc.want)
		}
		for _, n := range got {
			if n.Name == "python" && !reflect.DeepEqual(n.Projects, tc.project) {
				t.Errorf("%s: python projects %v, want %v", name, n.Projects, tc.project)
			}
		}
	}
}

func TestNeededKeepsAToolTiedToFilesForTheProjectsThatHoldThem(t *testing.T) {
	gradleDir, mavenDir := t.TempDir(), t.TempDir()
	for dir, file := range map[string]string{gradleDir: "build.gradle.kts", mavenDir: "pom.xml"} {
		if err := os.WriteFile(filepath.Join(dir, file), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	gradle := protocol.ToolNeed{Name: "gradle", Files: []string{"build.gradle.kts", "build.gradle"}}
	jvm := Source{Plugin: "jvm", Languages: []string{"java"}, Tools: []protocol.ToolNeed{tool("java", ""), gradle}}
	uses := []Use{{Project: "app", Language: "jvm", Root: gradleDir}, {Project: "lib", Language: "jvm", Root: mavenDir}}
	got := Needed("/w", nil, []Source{jvm}, uses, nil, nil)
	if want := []string{"java", "gradle"}; !reflect.DeepEqual(names(got), want) {
		t.Fatalf("got %v, want %v", names(got), want)
	}
	if want := []string{"app"}; !reflect.DeepEqual(got[1].Projects, want) {
		t.Errorf("gradle projects %v, want %v", got[1].Projects, want)
	}
	if got := Needed("/w", nil, []Source{jvm}, nil, []string{"java"}, nil); !reflect.DeepEqual(names(got), []string{"java"}) {
		t.Errorf("a language named by flag brought %v, want only java", names(got))
	}
}

func TestNeededAddsTheProjectFolderForPins(t *testing.T) {
	got := Needed("/w", nil, []Source{{Plugin: "node", Tools: []protocol.ToolNeed{tool("node", "")}}}, []Use{{Project: "web", Language: "node", Root: "/w/web"}}, nil, nil)
	if want := []string{"/w", "/w/web"}; !reflect.DeepEqual(got[0].Dirs, want) {
		t.Fatalf("dirs %v, want %v", got[0].Dirs, want)
	}
}

func TestLookup(t *testing.T) {
	found, unknown := Lookup("/w", []protocol.ToolNeed{tool("git", "")}, []Source{{Plugin: "go", Tools: []protocol.ToolNeed{tool("go", "")}}}, []string{"go", "git", "cobol"})
	if !reflect.DeepEqual(names(found), []string{"go", "git"}) || !reflect.DeepEqual(unknown, []string{"cobol"}) {
		t.Fatalf("found %v unknown %v", names(found), unknown)
	}
}
