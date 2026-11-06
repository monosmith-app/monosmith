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
		"en/a.json": {Data: []byte(`{"a.one": "1"}`)},
		"en/b.json": {Data: []byte(`{"b.one": "2"}`)},
		"pt/a.json": {Data: []byte(`{"a.one": "um"}`)},
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
		"en/a.json": {Data: []byte(`{"same": "1"}`)},
		"en/b.json": {Data: []byte(`{"same": "2"}`)},
	})
}

func TestReferenceIsACopyOfTheEnglishTexts(t *testing.T) {
	got := Reference()
	if len(got) == 0 || len(got) != len(catalogs[referenceLocale]) {
		t.Fatalf("Reference has %d keys, catalog has %d", len(got), len(catalogs[referenceLocale]))
	}
	got["test.copy"] = "x"
	if _, ok := catalogs[referenceLocale]["test.copy"]; ok {
		t.Fatal("changing the copy changed the catalog")
	}
}

func TestLoadCatalogReportsADuplicateKeyAsAnError(t *testing.T) {
	_, err := LoadCatalog(fstest.MapFS{
		"en/a.json": {Data: []byte(`{"same": "1"}`)},
		"en/b.json": {Data: []byte(`{"same": "2"}`)},
	})
	if err == nil {
		t.Fatal("a duplicate key must be an error")
	}
}
