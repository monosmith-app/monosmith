package workspace

import (
	"strings"
	"testing"
)

func TestADeclaredProjectNeedsARootALanguageAndAFolderThatIsClean(t *testing.T) {
	for input, want := range map[string]string{
		"[workspace]\nname = \"a\"\n[projects.mvd]\nlanguage = \"go\"\n":                                                                      "need root",
		"[workspace]\nname = \"a\"\n[projects.mvd]\nroot = \"apps/mvd\"\n":                                                                    "root needs language",
		"[workspace]\nname = \"a\"\n[projects.mvd]\nroot = \"../mvd\"\nlanguage = \"go\"\n":                                                   "workspace-relative folder",
		"[workspace]\nname = \"a\"\n[projects.mvd]\nroot = \".\"\nlanguage = \"go\"\n":                                                        "workspace-relative folder",
		"[workspace]\nname = \"a\"\n[projects.mvd]\nroot = 'apps\\mvd'\nlanguage = \"go\"\n":                                                  "forward slashes",
		"[workspace]\nname = \"a\"\n[projects.mvd]\nroot = \"apps/mvd\"\nlanguage = \"go\"\nkind = \"service\"\n":                             "kind must be",
		"[workspace]\nname = \"a\"\n[projects.a]\nroot = \"apps/x\"\nlanguage = \"go\"\n[projects.b]\nroot = \"apps/x\"\nlanguage = \"go\"\n": "is also the root of [projects.a]",
	} {
		if _, err := Parse([]byte(input)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("input %q: want error containing %q, got %v", input, want, err)
		}
	}
	cfg, err := Parse([]byte("[workspace]\nname = \"a\"\n[projects.mvd]\nroot = \"apps/mvd\"\nlanguage = \"go\"\nkind = \"app\"\n[projects.web]\ntags = [\"x\"]\n"))
	if err != nil || !cfg.Projects["mvd"].Declared() || cfg.Projects["web"].Declared() {
		t.Fatalf("%+v %v", cfg.Projects, err)
	}
}
