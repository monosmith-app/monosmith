package words

import (
	"testing"
	"testing/fstest"
)

func TestALocaleTagChoosesItsLanguageThenEnglish(t *testing.T) {
	catalogs["pt"] = map[string]string{}
	defer delete(catalogs, "pt")
	for tag, want := range map[string]string{"pt-BR": "pt", "pt_BR.UTF-8": "pt", "pt": "pt", "de": "en", "": "en", "C": "en"} {
		if got := match(tag); got != want {
			t.Errorf("match(%q) = %q, want %q", tag, got, want)
		}
	}
}

func TestALocaleThatLacksAKeyShowsTheEnglishText(t *testing.T) {
	catalogs["xx"] = map[string]string{"test.only.xx": "xx text"}
	catalogs["en"]["test.both"] = "english"
	defer delete(catalogs, "xx")
	defer delete(catalogs["en"], "test.both")
	Use("xx")
	defer Use("")
	if got := Say("test.only.xx"); got != "xx text" {
		t.Errorf("own key = %q", got)
	}
	if got := Say("test.both"); got != "english" {
		t.Errorf("fallback = %q", got)
	}
}

func TestTheCatalogLoadsEveryFileOfALocale(t *testing.T) {
	got := mustLoad(fstest.MapFS{
		"catalog/en/a.json": {Data: []byte(`{"a.one": "1"}`)},
		"catalog/en/b.json": {Data: []byte(`{"b.one": "2"}`)},
		"catalog/pt/a.json": {Data: []byte(`{"a.one": "um"}`)},
	})
	if got["en"]["a.one"] != "1" || got["en"]["b.one"] != "2" || got["pt"]["a.one"] != "um" {
		t.Fatalf("loaded %v", got)
	}
}

func TestAKeyDefinedInTwoFilesIsABuildError(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("a duplicate key must panic")
		}
	}()
	mustLoad(fstest.MapFS{
		"catalog/en/a.json": {Data: []byte(`{"same": "1"}`)},
		"catalog/en/b.json": {Data: []byte(`{"same": "2"}`)},
	})
}
