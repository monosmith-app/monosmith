package tomltext

import "testing"

const cargo = `[package]
name = "core"
publish = false # internal

[package.metadata.smith]
kind = "lib"

[dependencies]
`

func TestSetStringReplacesTheValueAndKeepsTheRest(t *testing.T) {
	got := SetString(cargo, "package.metadata.smith", "kind", "package")
	want := "[package]\nname = \"core\"\npublish = false # internal\n\n[package.metadata.smith]\nkind = \"package\"\n\n[dependencies]\n"
	if got != want {
		t.Fatalf("got:\n%s", got)
	}
}

func TestSetStringAddsMissingKeysAndTables(t *testing.T) {
	got := SetString("[package]\nname = \"x\"\n", "package.metadata.smith", "kind", "lib")
	if got != "[package]\nname = \"x\"\n\n[package.metadata.smith]\nkind = \"lib\"\n" {
		t.Fatalf("new table: %q", got)
	}
	got = SetString("[tool.smith]\nstack = \"s\"\n", "tool.smith", "kind", "lib")
	if got != "[tool.smith]\nstack = \"s\"\nkind = \"lib\"\n" {
		t.Fatalf("new key: %q", got)
	}
}

func TestRawAndRemove(t *testing.T) {
	if v, ok := Raw(cargo, "package", "publish"); !ok || v != "false" {
		t.Fatalf("raw: %q %v", v, ok)
	}
	out, ok := Remove(cargo, "package", "publish")
	if !ok || out != "[package]\nname = \"core\"\n\n[package.metadata.smith]\nkind = \"lib\"\n\n[dependencies]\n" {
		t.Fatalf("remove: %q", out)
	}
	if _, ok := Raw(cargo, "package", "kind"); ok {
		t.Fatal("kind is in another table")
	}
	crlf := "[package]\r\npublish = false\r\n"
	if got := SetRaw(crlf, "package", "edition", "\"2021\""); got != "[package]\r\npublish = false\r\nedition = \"2021\"\r\n" {
		t.Fatalf("crlf: %q", got)
	}
}
