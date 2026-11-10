package compose

import "testing"

func catalogueForSearch() []Item {
	return Build(Source{Offers: fakeOffers(), Languages: []Language{{Name: "node", Title: "TypeScript / JavaScript"}, {Name: "python", Title: "Python"}}}).Items
}

func names(hits []Hit) []string {
	var out []string
	for _, h := range hits {
		out = append(out, h.Language+"/"+h.Name)
	}
	return out
}

func TestSearchFindsByAliasAndSaysSo(t *testing.T) {
	hits := Search(catalogueForSearch(), "desktop")
	if len(hits) != 1 || hits[0].Name != "electron" || hits[0].MatchedBy != "alias" {
		t.Fatalf("desktop finds the shell by its alias: %v", names(hits))
	}
}

func TestSearchFindsServerlessHostsOfEveryLanguage(t *testing.T) {
	got := names(Search(catalogueForSearch(), "serverless"))
	if len(got) != 2 {
		t.Fatalf("both Azure Functions items (node has the alias): %v", got)
	}
}

func TestSearchMatchesTheStartOfAWordAndIgnoresCase(t *testing.T) {
	if hits := Search(catalogueForSearch(), "DESK"); len(hits) != 1 {
		t.Fatalf("a prefix finds it: %v", names(hits))
	}
	if hits := Search(catalogueForSearch(), "esktop"); len(hits) != 0 {
		t.Fatalf("a middle of a word does not: %v", names(hits))
	}
}

func TestSearchNeedsEveryWordAndRanksNameBeforeAlias(t *testing.T) {
	hits := Search(catalogueForSearch(), "react web")
	if len(hits) != 1 || hits[0].Name != "react" {
		t.Fatalf("both words match react only: %v", names(hits))
	}
	ranked := Search(catalogueForSearch(), "azure")
	if ranked[0].MatchedBy != "title" && ranked[0].MatchedBy != "name" {
		t.Fatalf("a name beats other matches: %+v", ranked[0])
	}
}

func TestAnEmptySearchListsEverythingInCatalogueOrder(t *testing.T) {
	items := catalogueForSearch()
	hits := Search(items, "  ")
	if len(hits) != len(items) || hits[0].ID != items[0].ID {
		t.Fatalf("nothing typed shows the whole catalogue: %d of %d", len(hits), len(items))
	}
}

func TestTheLanguageNameFindsItsChoices(t *testing.T) {
	got := names(Search(catalogueForSearch(), "python"))
	found := false
	for _, n := range got {
		if n == "python/azure-functions" {
			found = true
		}
	}
	if !found {
		t.Fatalf("a language word finds the items for that language: %v", got)
	}
}
