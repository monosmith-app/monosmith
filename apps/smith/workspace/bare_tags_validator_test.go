package workspace

import (
	"strings"
	"testing"
)

func TestBareTagsAreValidForOneProject(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"a\"\n[projects.java]\nbare_tags = true\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Projects["java"].BareTags {
		t.Errorf("bare_tags was not read: %+v", cfg.Projects["java"])
	}
}

func TestBareTagsAreRefusedWhereTheyWouldBeAmbiguous(t *testing.T) {
	for name, tc := range map[string]struct{ toml, want string }{
		"two projects":      {"[projects.a]\nbare_tags = true\n[projects.b]\nbare_tags = true\n", "a and b all set bare_tags"},
		"with a tag prefix": {"[projects.a]\nbare_tags = true\ntag_prefix = \"v\"\n", "bare_tags and tag_prefix"},
	} {
		_, err := Parse([]byte("[workspace]\nname = \"w\"\n" + tc.toml))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want %q", name, err, tc.want)
		}
	}
}

func TestRegistryNoneIsValidForReleasesAndProjectsOnly(t *testing.T) {
	for name, toml := range map[string]string{
		"release": "[release]\nregistry = \"none\"\n",
		"project": "[projects.java]\nregistry = \"none\"\n",
	} {
		if _, err := Parse([]byte("[workspace]\nname = \"w\"\n" + toml)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	for name, tc := range map[string]struct{ toml, want string }{
		"installs":        {"[workspace]\nname = \"w\"\nregistry = \"none\"\n", "installs always come from somewhere"},
		"a registry name": {"[workspace]\nname = \"w\"\n[registries.none]\nauth = \"token\"\ntoken_env = \"T\"\nnpm = \"https://x.test/\"\n", "is reserved"},
	} {
		_, err := Parse([]byte(tc.toml))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want %q", name, err, tc.want)
		}
	}
}
