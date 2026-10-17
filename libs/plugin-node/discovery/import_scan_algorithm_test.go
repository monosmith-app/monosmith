package discovery

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPackageNames(t *testing.T) {
	src := `
import { a } from '@acme/ui'
import b from "lodash/fp"
import './side-effect'
import type { C } from '@acme/ui/types'
export * from '@acme/core'
const d = require('express')
const e = await import('zod')
import fs from 'node:fs'
import x from '../local'
`
	got := PackageNames(src)
	want := []string{"@acme/ui", "lodash", "@acme/ui", "@acme/core", "express", "zod"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestPackageNamesSkipsWordsInsideStrings(t *testing.T) {
	// Both from OpenCraw: "from" and "import" as words in strings (#298).
	src := `
import { z } from 'zod'
throw new Error('the step has no "from" is given')
const label = word('from'), field = pillField(step, 'from', 'page')
const hint = "use import('x') to load it"
const t = ` + "`" + `copy from "a" to 'b'` + "`" + `
`
	if got, want := PackageNames(src), []string{"zod"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestScanImportsSkipsBuildOutput(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("src/main.ts", "import { x } from '@acme/ui'")
	write("dist/main.js", "require('@acme/stale')")
	write("node_modules/y/index.js", "require('@acme/other')")
	write("src/types.d.ts", "import '@acme/declared'")
	got, err := ScanImports(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"@acme/ui"}) {
		t.Fatalf("got %v", got)
	}
}

func TestPackageNamesSkipsComments(t *testing.T) {
	src := "// the old config did require('@nx/rollup/with-nx')\n" +
		"/**\n * `import ... from 'chromium-bidi/x'` failed\n */\n" +
		"const url = 'https://example.com/a' // trailing\n" +
		"const tpl = `// not a comment ${x}`\n" +
		"import { z } from 'zod' /* inline */ import y from \"yaml\"\n" +
		"const s = 'it\\'s' ; const r = require('lodash')\n"
	got := PackageNames(src)
	want := []string{"zod", "yaml", "lodash"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestScanFileImportsNamesEachFile(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("src/main.ts", "import { x } from '@acme/ui'\nimport { y } from '#internal/y'")
	write("src/util/fmt.ts", "import type { Z } from 'zod'")
	write("README.md", "import nothing from 'here'")
	got, err := ScanFileImports(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{"src/main.ts": {"@acme/ui"}, "src/util/fmt.ts": {"zod"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
