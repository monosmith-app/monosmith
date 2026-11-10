package adoptsplit

import "testing"

func TestRulesSayWhatStaysAtTheRoot(t *testing.T) {
	cases := []struct {
		eco, name string
		want      bool
	}{
		{Node, ".prettierrc", true},
		{Node, ".prettierrc.json", true},
		{Node, "Prettier.config.mjs", true},
		{Node, ".eslintrc.json", true},
		{Node, "eslint.config.mjs", true},
		{Node, "package-lock.json", true},
		{Node, "yarn.lock", true},
		{Node, "package.json", false},
		{Node, "tsconfig.json", false},
		{Node, "jest.config.js", false},
		{Node, "src", false},
		{Go, ".golangci.yml", true},
		{Go, "go.mod", false},
		{Go, "package-lock.json", false},
		{Python, "ruff.toml", true},
		{Python, "pyproject.toml", false},
		{Rust, "rustfmt.toml", true},
		{Rust, "rust-toolchain.toml", true},
		{Rust, "Cargo.toml", false},
		{"cobol", "anything", false},
	}
	for _, tc := range cases {
		if got := For(tc.eco).StaysAtRoot(tc.name); got != tc.want {
			t.Errorf("%s %s: StaysAtRoot = %v, want %v", tc.eco, tc.name, got, tc.want)
		}
	}
}

func TestWhatEachLanguageDeclares(t *testing.T) {
	if r := For(Go); r.Sort != nil || r.Split != nil {
		t.Error("go moves go.mod whole: adopt's workspace step writes go.work")
	}
	for _, eco := range []string{Node, Python} {
		if r := For(eco); r.Sort == nil || r.Split == nil {
			t.Errorf("%s must declare Sort and Split", eco)
		}
	}
	if r := For(Rust); r.Sort != nil || r.Split == nil {
		t.Error("rust declares a Split and no dependency Sort: dev-dependencies are per crate")
	}
}

func TestNodeToolsAndLintStack(t *testing.T) {
	for _, name := range []string{"typescript", "jest", "ts-jest", "@types/node", "@types/jest", "prettier", "prettier-plugin-x", "vitest", "@jest/globals", "husky"} {
		if !IsNodeTool(name) || IsNodeLint(name) {
			t.Errorf("%s must be a tool and not lint", name)
		}
	}
	for _, name := range []string{"eslint", "@typescript-eslint/parser", "eslint-plugin-react", "eslint-config-airbnb", "@eslint/js", "typescript-eslint"} {
		if !IsNodeLint(name) || IsNodeTool(name) {
			t.Errorf("%s must be lint and not a tool", name)
		}
	}
	for _, name := range []string{"supertest", "webpack", "nock", "typescript-plugin-x"} {
		if IsNodeTool(name) || IsNodeLint(name) {
			t.Errorf("%s is not known: the person is asked", name)
		}
	}
}
