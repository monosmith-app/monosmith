package recipereleases

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseReadsAReleaseUrl(t *testing.T) {
	for _, ref := range []string{
		"https://github.com/acme/recipes/releases/download/team@1.2.0/team.recipe.toml",
		"https://github.com/acme/recipes/releases/download/team%401.2.0/team.recipe.toml",
		"https://raw.githubusercontent.com/acme/recipes/team@1.2.0/recipes/team/team.recipe.toml",
		"https://raw.githubusercontent.com/acme/recipes/refs/tags/team@1.2.0/recipes/team/team.recipe.toml",
		"https://github.com/acme/recipes/raw/team%401.2.0/recipes/team/team.recipe.toml",
	} {
		loc, ok := Parse(ref)
		if !ok || loc.Owner != "acme" || loc.Repo != "recipes" || loc.Name != "team" || loc.Version != "1.2.0" {
			t.Errorf("%s: %+v %v", ref, loc, ok)
			continue
		}
		if got := loc.Pinned("1.3.0"); !strings.Contains(got, "1.3.0") || strings.Contains(got, "1.2.0") || strings.Replace(got, "1.3.0", "1.2.0", 1) != ref {
			t.Errorf("%s: pinned = %s", ref, got)
		}
	}
	for _, ref := range []string{
		"https://github.com/acme/recipes/raw/main/recipes/team/team.recipe.toml",
		"https://raw.githubusercontent.com/acme/recipes/v1.2.0/team.recipe.toml",
		"https://example.com/team@1.2.0/team.recipe.toml",
		"http://github.com/acme/recipes/releases/download/team@1.2.0/team.recipe.toml",
		"./team.recipe.toml",
	} {
		if loc, ok := Parse(ref); ok {
			t.Errorf("%s is not a pinned release, but parsed as %+v", ref, loc)
		}
	}
}

func TestLatestFindsTheNewestReleaseOfTheRecipe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/acme/recipes/releases" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[
			{"tag_name": "team@1.9.0"}, {"tag_name": "team@1.10.0"}, {"tag_name": "team@2.0.0", "prerelease": true},
			{"tag_name": "team@3.0.0", "draft": true}, {"tag_name": "company@5.0.0"}, {"tag_name": "v9.9.9"}
		]`))
	}))
	defer server.Close()
	APIURL = server.URL
	t.Cleanup(func() { APIURL = "https://api.github.com" })
	loc, _ := Parse("https://github.com/acme/recipes/releases/download/team@1.2.0/team.recipe.toml")
	got, err := Latest(server.Client(), loc)
	if err != nil || got != "1.10.0" {
		t.Errorf("latest = %q, %v; 1.10 is newer than 1.9, and drafts, prereleases and other recipes don't count", got, err)
	}
	other, _ := Parse("https://github.com/acme/recipes/releases/download/unknown@1.0.0/x.recipe.toml")
	if _, err := Latest(server.Client(), other); err == nil {
		t.Error("a recipe the repository never released has a latest version")
	}
	missing, _ := Parse("https://github.com/acme/nothing/releases/download/team@1.2.0/team.recipe.toml")
	if _, err := Latest(server.Client(), missing); err == nil {
		t.Error("a missing repository has a latest version")
	}
}

func TestNewerComparesNumbers(t *testing.T) {
	if !Newer("1.10.0", "1.9.9") || Newer("1.2.0", "1.2.0") || Newer("1.2.0", "2.0.0") {
		t.Error("versions are compared as numbers")
	}
}
