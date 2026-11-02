package tauriadd

import (
	"bufio"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
)

func TestDefaultIdentifierIsLoweredAndKeepsHyphens(t *testing.T) {
	for _, tc := range []struct{ ws, project, want string }{
		{"acme", "studio", "com.acme.studio"},
		{"Acme Corp", "My_Studio", "com.acmecorp.mystudio"},
		{"", "---", "com.app.---"},
	} {
		if got := DefaultIdentifier(tc.ws, tc.project); got != tc.want {
			t.Errorf("%q %q: got %q, want %q", tc.ws, tc.project, got, tc.want)
		}
	}
}

func TestValidIdentifier(t *testing.T) {
	for id, ok := range map[string]bool{
		"com.acme.studio": true, "io.github.me-and-you.app": true, "a.b": true,
		"studio": false, "com..studio": false, "com.1acme.app": false, "com.acme.my app": false, "": false,
	} {
		if err := ValidIdentifier(id); (err == nil) != ok {
			t.Errorf("%q: valid=%v, error %v", id, ok, err)
		}
	}
}

func TestSetIdentifierChangesOnlyTheField(t *testing.T) {
	conf := "{\n  \"productName\": \"studio\",\n  \"identifier\": \"com.acme.studio\",\n  \"version\": \"0.1.0\"\n}\n"
	got, ok := SetIdentifier(conf, "com.real.app")
	if !ok || !strings.Contains(got, `"identifier": "com.real.app"`) || !strings.Contains(got, `"productName": "studio"`) || strings.Contains(got, "com.acme") {
		t.Fatalf("%v\n%s", ok, got)
	}
	if _, ok := SetIdentifier(`{"name":"x"}`, "com.real.app"); ok {
		t.Error("a file without the field is left alone")
	}
}

func TestPlanForListsTheFilesAndSkipsAnExistingSrcTauri(t *testing.T) {
	var b strings.Builder
	operationplan.Render(&b, PlanFor("studio", "apps/studio", "react-tauri", false, "com.acme.studio"))
	for _, want := range []string{"apps/studio/src-tauri/Cargo.toml", "com.acme.studio", "smith.toml", "the CI pipelines", "install"} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("the plan lacks %q:\n%s", want, b.String())
		}
	}
	b.Reset()
	operationplan.Render(&b, PlanFor("studio", "apps/studio", "react-tauri", true, ""))
	if strings.Contains(b.String(), "Cargo.toml") {
		t.Errorf("an existing src-tauri is left as it is:\n%s", b.String())
	}
}

func TestAskIdentifier(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"enter takes the default", "\n", "com.acme.studio"},
		{"a typed identifier", "io.me.app\n", "io.me.app"},
		{"a wrong one is asked again", "nope\nio.me.app\n", "io.me.app"},
		{"closed input takes the default", "", "com.acme.studio"},
	} {
		var out strings.Builder
		if got := askIdentifier(bufio.NewReader(strings.NewReader(tc.input)), &out, "com.acme.studio"); got != tc.want {
			t.Errorf("%s: got %q\n%s", tc.name, got, out.String())
		}
	}
}
