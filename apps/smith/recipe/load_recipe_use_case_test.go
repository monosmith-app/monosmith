package recipe

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func writeFile(t *testing.T, dir, rel, body string) string {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFollowsExtendsRelativeToTheFileThatHasIt(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "company/base.toml", "name = \"company\"\nversion = \"3.0.0\"\n")
	writeFile(t, dir, "teams/mid.toml", "name = \"mid\"\nextends = \"../company/base.toml\"\n")
	top := writeFile(t, dir, "teams/top.toml", "name = \"top\"\nextends = \"mid.toml\"\n")
	loaded, err := (&Loader{}).Load(top)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, d := range loaded.Chain {
		names = append(names, d.Recipe.Name)
	}
	if strings.Join(names, ",") != "company,mid,top" {
		t.Errorf("chain = %v, the oldest ancestor must come first", names)
	}
}

func TestLoadRefusesWhatItCannotChain(t *testing.T) {
	dir := t.TempDir()
	a := writeFile(t, dir, "a.toml", "name = \"a\"\nextends = \"b.toml\"\n")
	writeFile(t, dir, "b.toml", "name = \"b\"\nextends = \"a.toml\"\n")
	orphan := writeFile(t, dir, "orphan.toml", "name = \"o\"\n\nextends = \"nowhere.toml\"\n")
	insecure := writeFile(t, dir, "insecure.toml", "name = \"i\"\nextends = \"http://example.com/r.toml\"\n")
	other := writeFile(t, dir, "other.toml", "name = \"x\"\nextends = \"git://example.com/r.toml\"\n")
	for name, tc := range map[string]struct{ ref, want string }{
		"cycle":          {a, "recipes extend each other in a cycle: " + a},
		"missing parent": {orphan, "orphan.toml:3: extends \"nowhere.toml\": reading the recipe"},
		"missing file":   {filepath.Join(dir, "none.toml"), "reading the recipe"},
		"plain http":     {insecure, "recipes load over https only"},
		"other scheme":   {other, "a recipe is a path or an https URL"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := (&Loader{}).Load(tc.ref)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestLoadStopsAnUnreasonablyLongChain(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 20; i++ {
		body := "name = \"r\"\n"
		if i < 19 {
			body += "extends = \"r" + string(rune('a'+i+1)) + ".toml\"\n"
		}
		writeFile(t, dir, "r"+string(rune('a'+i))+".toml", body)
	}
	_, err := (&Loader{}).Load(filepath.Join(dir, "ra.toml"))
	if err == nil || !strings.Contains(err.Error(), "recipes extend one another") {
		t.Errorf("err = %v", err)
	}
}

// recipeServer serves a recipe whose body can change, and counts requests.
func recipeServer(t *testing.T, body *atomic.Value, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch r.URL.Path {
		case "/team.toml":
			_, _ = w.Write([]byte(body.Load().(string)))
		case "/parent/base.toml":
			_, _ = w.Write([]byte("name = \"base\"\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestLoadPinsAUrlRecipeAndRefusesItsChange(t *testing.T) {
	var body atomic.Value
	var hits atomic.Int32
	body.Store("name = \"team\"\nversion = \"1.0.0\"\n")
	srv := recipeServer(t, &body, &hits)
	root := t.TempDir()
	url := srv.URL + "/team.toml"
	loader := &Loader{Root: root, Client: srv.Client()}

	loaded, err := loader.Load(url)
	if err != nil || loaded.Chain[0].Recipe.Version != "1.0.0" {
		t.Fatalf("first load: %v", err)
	}
	lock, err := os.ReadFile(filepath.Join(root, LockFile))
	if err != nil || !strings.Contains(string(lock), url) || !strings.Contains(string(lock), "sha256 = ") {
		t.Fatalf("the load must pin the recipe in %s:\n%s (%v)", LockFile, lock, err)
	}
	if _, err := os.Stat(cachePath(root, url)); err != nil {
		t.Fatalf("the load must cache the recipe: %v", err)
	}

	// The URL's content changes; the pinned, cached copy is still used, with
	// no request at all, so a build does not change under you.
	body.Store("name = \"team\"\nversion = \"2.0.0\"\n")
	before := hits.Load()
	loaded, err = loader.Load(url)
	if err != nil || loaded.Chain[0].Recipe.Version != "1.0.0" || hits.Load() != before {
		t.Fatalf("a pinned recipe must come from the cache: %v, version %s, %d requests", err, loaded.Chain[0].Recipe.Version, hits.Load()-before)
	}

	// On a machine without the cache (a fresh clone), the changed content is
	// refused against the lockfile.
	if err := os.RemoveAll(filepath.Join(root, CacheDir)); err != nil {
		t.Fatal(err)
	}
	_, err = loader.Load(url)
	var changed *ChangedError
	if !errors.As(err, &changed) || changed.URL != url || !strings.Contains(err.Error(), "smith upgrade") {
		t.Fatalf("err = %v, want a ChangedError that points at smith upgrade", err)
	}

	// Accepting the change pins the new content.
	change, err := loader.Accept(url)
	if err != nil || !change.Changed() || change.Old == "" || change.Old == change.New {
		t.Fatalf("accept: %+v, %v", change, err)
	}
	loaded, err = loader.Load(url)
	if err != nil || loaded.Chain[0].Recipe.Version != "2.0.0" {
		t.Fatalf("after accepting: %v", err)
	}
	if again, err := loader.Accept(url); err != nil || again.Changed() {
		t.Errorf("accepting unchanged content must change nothing: %+v, %v", again, err)
	}
}

func TestLoadWithoutAWorkspaceFetchesAndPinsNothing(t *testing.T) {
	var body atomic.Value
	var hits atomic.Int32
	body.Store("name = \"team\"\n")
	srv := recipeServer(t, &body, &hits)
	loader := &Loader{Client: srv.Client()}
	for i := 0; i < 2; i++ {
		if _, err := loader.Load(srv.URL + "/team.toml"); err != nil {
			t.Fatal(err)
		}
	}
	if hits.Load() != 2 {
		t.Errorf("%d requests, want one per load", hits.Load())
	}
	if _, err := loader.Accept(srv.URL + "/team.toml"); err == nil {
		t.Error("accepting needs a workspace to pin in")
	}
}

func TestLoadResolvesExtendsOfAUrlRecipeAgainstItsUrl(t *testing.T) {
	var body atomic.Value
	var hits atomic.Int32
	body.Store("name = \"team\"\nextends = \"parent/base.toml\"\n")
	srv := recipeServer(t, &body, &hits)
	loaded, err := (&Loader{Root: t.TempDir(), Client: srv.Client()}).Load(srv.URL + "/team.toml")
	if err != nil || len(loaded.Chain) != 2 || loaded.Chain[0].Recipe.Name != "base" || !loaded.Chain[0].Source.URL {
		t.Fatalf("chain = %+v, %v", loaded.Chain, err)
	}
}

func TestLoadReportsAUrlThatAnswersWithAnError(t *testing.T) {
	var body atomic.Value
	var hits atomic.Int32
	srv := recipeServer(t, &body, &hits)
	_, err := (&Loader{Client: srv.Client()}).Load(srv.URL + "/missing.toml")
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("err = %v", err)
	}
}
