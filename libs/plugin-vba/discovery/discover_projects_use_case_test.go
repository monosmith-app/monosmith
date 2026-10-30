package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("apps/report/main.bas", "Attribute VB_Name = \"Main\"\r\n' smithKind: app\r\n' smithVersion: 1.2.0\r\nOption Explicit\r\n")
	write("apps/report/Totals.cls", "Attribute VB_Name = \"Totals\"\r\n")
	write("libs/strings/Alpha.cls", "Attribute VB_Name = \"Alpha\"\r\n' smithVersion: 9.9.9\r\n")
	write("libs/strings/Zeta.bas", "Attribute VB_Name = \"Zeta\"\r\n' smithStack: office\r\n' smithVersion: 0.3.0\r\n")
	write("docs/paper/thesis.cls", "\\NeedsTeXFormat{LaTeX2e}\n")
	write("legacy.bas", "Attribute VB_Name = \"Legacy\"\r\n")
	res, err := Discover(ws, []string{"legacy.bas", "apps/report/Totals.cls", "apps/report/main.bas", "libs/strings/Alpha.cls", "libs/strings/Zeta.bas", "docs/paper/thesis.cls"})
	if err != nil || len(res.Projects) != 2 {
		t.Fatal(err, res.Projects)
	}
	report, strs := res.Projects[0], res.Projects[1]
	if report.Name != "report" || report.Kind != protocol.ProjectApp || report.Version != "1.2.0" || report.Exports[0] != "apps/report" || len(report.Imports) != 0 || len(report.Tasks) != 0 {
		t.Errorf("report %+v", report)
	}
	if strs.Kind != protocol.ProjectLib || strs.Stack != "office" || strs.Version != "0.3.0" {
		t.Errorf("a .bas module is the main one before a .cls: %+v", strs)
	}
}

func TestFindMain(t *testing.T) {
	files := map[string]string{
		"b.bas": "Attribute VB_Name = \"B\"\r\n", "a.cls": "Attribute VB_Name = \"A\"\r\n", "c.frm": "VERSION 5.00\r\nAttribute VB_Name = \"C\"\r\n",
		"main.bas": "not a module", "notes.txt": "",
	}
	read := func(n string) (string, bool) { s, ok := files[n]; return s, ok }
	names := []string{"notes.txt", "c.frm", "a.cls", "b.bas", "main.bas"}
	if got, _ := FindMain(names, read); got != "b.bas" {
		t.Errorf("main.bas is not a VBA module, so the first .bas is: %q", got)
	}
	delete(files, "b.bas")
	if got, _ := FindMain(names, read); got != "a.cls" {
		t.Errorf("a .cls comes before a .frm: %q", got)
	}
	if got, _ := FindMain([]string{"notes.txt"}, read); got != "" {
		t.Errorf("no module, no main: %q", got)
	}
}

func TestSetField(t *testing.T) {
	out, ok := SetField("Attribute VB_Name = \"M\"\r\nOption Explicit\r\n", "Version", "1.0.0")
	if !ok || out != "Attribute VB_Name = \"M\"\r\n' smithVersion: 1.0.0\r\nOption Explicit\r\n" {
		t.Errorf("a new comment goes under the Attribute line, in the file's line ending: %q", out)
	}
	if again, ok := SetField(out, "Version", "1.0.0"); ok || again != out {
		t.Error("the same version is no change")
	}
	if out, _ := SetField(out, "Version", "2.0.0"); out != "Attribute VB_Name = \"M\"\r\n' smithVersion: 2.0.0\r\nOption Explicit\r\n" {
		t.Errorf("%q", out)
	}
	if out, _ := SetField("Option Explicit\n", "Kind", "lib"); out != "' smithKind: lib\nOption Explicit\n" {
		t.Errorf("without an Attribute line it goes first: %q", out)
	}
}
