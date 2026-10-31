package jsondoc

import (
	"strings"
	"testing"
)

func TestDocumentKeepsOrderAndEditsInPlace(t *testing.T) {
	src := `{
  "name": "@acme/web",
  "version": "1.0.0",
  "nx": {
    "targets": {
      "build": { "executor": "x" }
    }
  },
  "files": ["dist", "!**/*.map"],
  "scripts": {
    "test": "jest",
    "e2e": "a && b <c>"
  },
  "count": 1.50
}
`
	doc, err := ParseDocument([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	doc.Delete("nx")
	doc.Object("scripts", false).SetString("build", "tsc -p tsconfig.lib.json")
	doc.Object("scripts", false).SetString("test", "jest --ci")
	doc.Object("smith", true).SetString("kind", "package")
	want := `{
  "name": "@acme/web",
  "version": "1.0.0",
  "files": [
    "dist",
    "!**/*.map"
  ],
  "scripts": {
    "test": "jest --ci",
    "e2e": "a && b <c>",
    "build": "tsc -p tsconfig.lib.json"
  },
  "count": 1.50,
  "smith": {
    "kind": "package"
  }
}
`
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	if _, err := ParseDocument([]byte("[1]")); err == nil || !strings.Contains(err.Error(), "object") {
		t.Fatalf("an array is not a document: %v", err)
	}
	empty, _ := ParseDocument([]byte(`{"a": {}}`))
	if string(empty.Bytes()) != "{\n  \"a\": {}\n}\n" {
		t.Fatalf("%q", empty.Bytes())
	}
}

func TestSetJSONMakesObjectsEditable(t *testing.T) {
	doc, _ := ParseDocument([]byte(`{"name": "x"}`))
	if err := doc.SetJSON("jest", map[string]any{"preset": "ts-jest", "roots": []string{"<rootDir>/src"}}); err != nil {
		t.Fatal(err)
	}
	doc.Object("jest", false).SetString("testEnvironment", "node")
	if err := doc.SetJSON("private", true); err != nil {
		t.Fatal(err)
	}
	want := `{
  "name": "x",
  "jest": {
    "preset": "ts-jest",
    "roots": [
      "<rootDir>/src"
    ],
    "testEnvironment": "node"
  },
  "private": true
}
`
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestDocumentKeepsCRLF(t *testing.T) {
	doc, err := ParseDocument([]byte("{\r\n  \"name\": \"a\"\r\n}\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	doc.SetString("version", "1.0.0")
	want := "{\r\n  \"name\": \"a\",\r\n  \"version\": \"1.0.0\"\r\n}\r\n"
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestString(t *testing.T) {
	doc, _ := ParseDocument([]byte(`{"a": "x", "n": 1, "o": {}}`))
	if v, ok := doc.String("a"); !ok || v != "x" {
		t.Fatalf("%q %v", v, ok)
	}
	for _, k := range []string{"n", "o", "missing"} {
		if _, ok := doc.String(k); ok {
			t.Errorf("%s is not a string", k)
		}
	}
}
