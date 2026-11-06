package main

import (
	"strings"
	"testing"
)

func TestKotlinHasATypedConstantPerKeyAndTheSayHelper(t *testing.T) {
	got, err := Kotlin(fixture(t, "jetbrains"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"package dev.monosmith.jetbrains.words",
		"object SmithWordsConfig {",
		"/** The key cmd.remove.title. Values: {0} is name, {1} is project. */\n    const val CMD_REMOVE_TITLE = \"cmd.remove.title\"",
		"const val CMD_ITEMS_ONE = \"cmd.items.one\"",
		"fun say(key: String, vararg values: Any): String",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
}

func TestKotlinOfNoKeysStillHasSay(t *testing.T) {
	got, err := Kotlin(Resolved{Texts: map[string]map[string]Entry{"en": {}}})
	if err != nil || !strings.Contains(got, "fun say(") || strings.Contains(got, "const val CMD") {
		t.Fatalf("got %v\n%s", err, got)
	}
}

func TestTwoKeysWithOneConstantNameAreAnError(t *testing.T) {
	r := Resolved{Keys: []string{"a.b-c", "a.b.c"}, Texts: map[string]map[string]Entry{"en": {"a.b-c": {Text: "x"}, "a.b.c": {Text: "y"}}}}
	if _, err := Kotlin(r); err == nil {
		t.Fatal("the same constant twice must be an error")
	}
}

func TestConstantNameIsUpperSnakeAndNeverStartsWithADigit(t *testing.T) {
	for key, want := range map[string]string{"cmd.add-title": "CMD_ADD_TITLE", "9lives": "_9LIVES"} {
		if got := ConstantName(key); got != want {
			t.Errorf("ConstantName(%q) = %q, want %q", key, got, want)
		}
	}
}
