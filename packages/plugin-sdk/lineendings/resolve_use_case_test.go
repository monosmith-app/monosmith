package lineendings

import "testing"

func TestResolveKeepsTheExistingEnding(t *testing.T) {
	dir := repo(t, "core.autocrlf", "input")
	put(t, dir, "ci.yml", "a: 1\r\nb: 2\r\n")
	if got := Resolve(dir, "ci.yml", "a: 1\nb: 3\n"); got != "a: 1\r\nb: 3\r\n" {
		t.Errorf("got %q", got)
	}
	put(t, dir, "lf.yml", "a: 1\nb: 2\n")
	if got := Resolve(dir, "lf.yml", "a: 1\r\nb: 3\r\n"); got != "a: 1\nb: 3\n" {
		t.Errorf("got %q", got)
	}
}

func TestResolveUsesThePolicyForANewFile(t *testing.T) {
	dir := repo(t, "core.autocrlf", "true")
	if got := Resolve(dir, "new.json", "{\n}\n"); got != "{\r\n}\r\n" {
		t.Errorf("got %q", got)
	}
}
