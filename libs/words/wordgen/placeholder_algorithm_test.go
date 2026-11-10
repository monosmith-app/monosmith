package main

import (
	"slices"
	"testing"
)

func TestPlaceholdersAreListedInOrderOfFirstAppearance(t *testing.T) {
	for text, want := range map[string][]string{
		"Remove {name} from {project}": {"name", "project"},
		"{b} {a} {b}":                  {"b", "a"},
		"no placeholder, {Upper} {1}":  nil,
	} {
		if got := Placeholders(text); !slices.Equal(got, want) {
			t.Errorf("Placeholders(%q) = %v, want %v", text, got, want)
		}
	}
}

func TestPositionalFollowsTheOrderOfTheEnglishText(t *testing.T) {
	order := []string{"name", "project"}
	for text, want := range map[string]string{
		"Remove {name} from {project}": "Remove {0} from {1}",
		"Remover {project} de {name}":  "Remover {1} de {0}",
		"{name} and {name}":            "{0} and {0}",
		"{other} stays":                "{other} stays",
	} {
		if got := Positional(text, order); got != want {
			t.Errorf("Positional(%q) = %q, want %q", text, got, want)
		}
	}
}
