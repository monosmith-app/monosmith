package swapchoice

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
)

func put(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func jestSetup() tooling.Setup {
	return tooling.Setup{
		DevDependencies: map[string]string{"jest": "^30", "ts-jest": "^29"},
		Scripts:         map[string]string{"test": "jest"},
		Fields:          map[string]any{"jest": map[string]any{"testEnvironment": "node"}},
		Files:           map[string]string{"jest.config.cts": "module.exports = {}\n"},
	}
}

func vitestSetup() tooling.Setup {
	return tooling.Setup{DevDependencies: map[string]string{"vitest": "^5"}, Scripts: map[string]string{"test": "vitest run"}}
}

func TestPlanMigrationRemovesWhatTheOldRunnerWroteAndKeepsEdits(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"devDependencies":{"jest":"^30","ts-jest":"^29","prettier":"^3"}}`)
	put(t, root, "apps/web/package.json", `{"scripts":{"test":"jest","build":"tsc"},"jest":{"testEnvironment":"node"}}`)
	put(t, root, "apps/web/jest.config.cts", "module.exports = { edited: true }\n")
	put(t, root, "apps/web/src/a.test.ts", "import { describe, expect, it } from '@jest/globals'\n\nit('x', () => expect(1).toBe(1))\n")
	put(t, root, "apps/web/src/b.test.ts", "import { jest, it } from '@jest/globals'\n\nit('x', () => jest.fn())\n")
	m, err := PlanMigration(root, "apps/web", jestSetup(), vitestSetup(), "@jest/globals", "vitest", false)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"jest", "ts-jest"}; !reflect.DeepEqual(m.Dependencies, want) {
		t.Errorf("dependencies %v, want %v", m.Dependencies, want)
	}
	if !reflect.DeepEqual(m.Scripts, []string{"test"}) || !reflect.DeepEqual(m.Fields, []string{"jest"}) {
		t.Errorf("scripts %v fields %v", m.Scripts, m.Fields)
	}
	if len(m.Files) != 0 {
		t.Errorf("an edited config must stay, got %v", m.Files)
	}
	if got := m.Rewrites["src/a.test.ts"]; !strings.HasPrefix(got, "import { describe, expect, it } from 'vitest'") {
		t.Errorf("a.test.ts not moved: %q", got)
	}
	if _, moved := m.Rewrites["src/b.test.ts"]; moved {
		t.Error("b.test.ts uses jest.fn and must be left")
	}
	joined := strings.Join(m.ByHand, "\n")
	for _, want := range []string{"jest.config.cts has your edits", "src/b.test.ts uses jest."} {
		if !strings.Contains(joined, want) {
			t.Errorf("by hand lacks %q:\n%s", want, joined)
		}
	}
}

func TestPlanMigrationKeepsRootPackagesAnotherProjectUses(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"devDependencies":{"jest":"^30","ts-jest":"^29"}}`)
	put(t, root, "apps/web/package.json", `{"scripts":{"test":"jest"}}`)
	m, err := PlanMigration(root, "apps/web", jestSetup(), vitestSetup(), "@jest/globals", "vitest", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Dependencies) != 0 || !reflect.DeepEqual(m.Kept, []string{"jest", "ts-jest"}) {
		t.Errorf("dependencies %v kept %v", m.Dependencies, m.Kept)
	}
}

func TestPlanMigrationDeletesAConfigFileTheOldRunnerWrote(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{}`)
	put(t, root, "apps/web/package.json", `{}`)
	put(t, root, "apps/web/jest.config.cts", "module.exports = {}\n")
	m, err := PlanMigration(root, "apps/web", jestSetup(), vitestSetup(), "@jest/globals", "vitest", false)
	if err != nil || !reflect.DeepEqual(m.Files, []string{"jest.config.cts"}) {
		t.Fatalf("files %v, err %v", m.Files, err)
	}
}
