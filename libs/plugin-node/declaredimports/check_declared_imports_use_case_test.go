package declaredimports

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A workspace like the one #284 came from: ocr uses core's types without
// declaring core, and its tests use tools the root declares.
func workspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, root, "package.json", `{"private": true, "workspaces": ["packages/*"], "devDependencies": {"jest": "^30.0.0", "typescript": "^6.0.0"}}`)
	write(t, root, "packages/core/package.json", `{"name": "@acme/core", "version": "1.0.0"}`)
	write(t, root, "packages/ocr/package.json", `{
  "name": "@acme/ocr",
  "version": "1.0.0",
  "dependencies": {"tesseract.js": "^6.0.0"},
  "devDependencies": {"@types/node": "^24.0.0", "zod": "^4.0.0"}
}`)
	write(t, root, "packages/ocr/src/recognise/recognise_text_use_case.ts", `
import { createWorker } from 'tesseract.js'
import type { Page } from '@acme/core'
import { readFile } from 'node:fs/promises'
import { join } from 'path'
import { schema } from 'zod'
import { local } from './local.js'
import { own } from '@acme/ocr/internal'
import { mapped } from '#internal/mapped'
`)
	write(t, root, "packages/ocr/src/recognise/recognise_text_use_case.test.ts", `
import { describe } from 'jest'
import { schema } from 'zod'
import { page } from '@acme/core'
`)
	write(t, root, "packages/ocr/jest.config.ts", `import type { Config } from 'jest'`)
	write(t, root, "packages/ocr/dist/index.js", `require('@acme/stale')`)
	return root
}

func TestCheckReportsWhatAShippedFileImportsWithoutDeclaring(t *testing.T) {
	got, err := Check(workspace(t), "packages/ocr")
	if err != nil {
		t.Fatal(err)
	}
	want := []UndeclaredImport{
		{File: "src/recognise/recognise_text_use_case.test.ts", Package: "@acme/core"},
		{File: "src/recognise/recognise_text_use_case.ts", Package: "@acme/core"},
		{File: "src/recognise/recognise_text_use_case.ts", Package: "zod", DevOnly: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
	if s := got[2].String(); s != "src/recognise/recognise_text_use_case.ts imports zod, declared only in devDependencies" {
		t.Errorf("String() = %q", s)
	}
}

func TestCheckPassesWhenThePeerIsDeclared(t *testing.T) {
	root := workspace(t)
	write(t, root, "packages/ocr/package.json", `{
  "name": "@acme/ocr",
  "dependencies": {"tesseract.js": "^6.0.0", "zod": "^4.0.0"},
  "peerDependencies": {"@acme/core": "^1.0.0"}
}`)
	got, err := Check(root, "packages/ocr")
	if err != nil || len(got) != 0 {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestCheckTakesAnOptionalPeerListedOnlyInItsMeta(t *testing.T) {
	root := t.TempDir()
	write(t, root, "package.json", `{"private": true}`)
	write(t, root, "packages/cfg/package.json", `{"name": "cfg", "peerDependenciesMeta": {"@nx/eslint-plugin": {"optional": true}, "rollup": {"optional": false}}}`)
	write(t, root, "packages/cfg/index.js", "let nx\ntry { nx = require('@nx/eslint-plugin') } catch {}\nimport r from 'rollup'\n")
	got, err := Check(root, "packages/cfg")
	if err != nil || len(got) != 1 || got[0].Package != "rollup" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestCheckLetsAVSCodeExtensionImportVSCode(t *testing.T) {
	root := t.TempDir()
	write(t, root, "package.json", `{"private": true}`)
	write(t, root, "apps/ext/package.json", `{"name": "ext", "engines": {"vscode": "^1.100.0"}}`)
	write(t, root, "apps/ext/src/main.ts", `import * as vscode from 'vscode'`)
	got, err := Check(root, "apps/ext")
	if err != nil || len(got) != 0 {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestDevFile(t *testing.T) {
	for rel, want := range map[string]bool{
		"src/a/a_use_case.ts":      false,
		"src/a/a_use_case.test.ts": true,
		"src/a/a.spec.tsx":         true,
		"e2e/smoke.e2e.ts":         true,
		"tools/gen.mjs":            true,
		"src/a/testing/a.mock.ts":  true,
		"jest.config.ts":           true,
		"vitest.config.mts":        true,
		"src/config/settings.ts":   false,
		"src/latest/attest.ts":     false,
	} {
		if got := DevFile(rel); got != want {
			t.Errorf("DevFile(%q) = %v, want %v", rel, got, want)
		}
	}
}

func TestCheckSkipsTheImportsATsconfigPathsAliasMaps(t *testing.T) {
	root := t.TempDir()
	write(t, root, "apps/api/package.json", `{"name": "api", "version": "1.0.0"}`)
	write(t, root, "apps/api/tsconfig.base.json", `{"compilerOptions": {"paths": {"@integrations": ["./src/integrations/index.ts"]}}}`)
	write(t, root, "apps/api/tsconfig.json", `{
  // comments, as tsconfig allows
  "extends": "./tsconfig.base.json",
  "compilerOptions": {"paths": {"@shared/*": ["./src/shared/*/index.ts"], "@integrations/*": ["./src/integrations/*"]}}
}`)
	write(t, root, "apps/api/src/features/send.use-case.ts", `
import { required } from '@shared/validation'
import type { WatchData } from '@shared/contracts'
import { jira } from '@integrations'
import { email } from '@integrations/email'
import { logger } from '@auto/logger'
`)
	got, err := Check(root, "apps/api")
	if err != nil {
		t.Fatal(err)
	}
	want := []UndeclaredImport{{File: "src/features/send.use-case.ts", Package: "@auto/logger"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want only the package no alias maps: %+v", got, want)
	}
}
