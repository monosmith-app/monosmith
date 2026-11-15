package turboplan

import (
	"strings"
	"testing"
)

func TestRewriteScript(t *testing.T) {
	names := map[string]string{"web": "web", "@acme/ui": "ui", "ui": "ui"}
	cases := []struct {
		in, want string
	}{
		{"turbo run build", "smith run build --bail"},
		{"turbo build test", "smith run build test --bail"},
		{"turbo run build --filter=web --continue", "smith run build -p web"},
		{"turbo run test --filter @acme/ui --filter=!web", "smith run test -p ui,!web --bail"},
		{"turbo run lint --affected --force --concurrency=4", "smith run lint --affected --no-cache --parallel 4 --bail"},
		{"turbo run build --ui=stream --output-logs=new-only", "smith run build --bail"},
		{"pnpm install", "smith install"},
		{"pnpm install --frozen-lockfile", "npm ci"},
		{"yarn install --immutable", "npm ci"},
		{"yarn", "smith install"},
		{"pnpm build", "npm run build"},
		{"pnpm run test --coverage", "npm run test -- --coverage"},
		{"yarn lint", "npm run lint"},
		{"pnpm turbo run build", "smith run build --bail"},
		{"yarn turbo build", "smith run build --bail"},
		{"npx turbo run build", "smith run build --bail"},
		{"pnpm -r run build", "smith run build"},
		{"pnpm --filter web build", "smith run build -p web"},
		{"pnpm --filter=@acme/ui run test", "smith run test -p ui"},
		{"yarn workspace web build", "smith run build -p web"},
		{"pnpm exec tsc --noEmit", "npx tsc --noEmit"},
		{"pnpm dlx create-foo", "npx create-foo"},
		{"CI=1 turbo run build", "CI=1 smith run build --bail"},
		{"rimraf dist && pnpm build", "rimraf dist && npm run build"},
		{"tsc --noEmit", "tsc --noEmit"},
	}
	for _, c := range cases {
		got := RewriteScript(c.in, names)
		if got.Problem != "" || got.Script != c.want {
			t.Errorf("%q\n got  %q (%s)\n want %q", c.in, got.Script, got.Problem, c.want)
		}
	}
	if got := RewriteScript("tsc --noEmit", names); got.Changed {
		t.Error("a script without the three tools is untouched")
	}
	if got := RewriteScript("turbo run build --ui=stream", names); len(got.Dropped) != 1 || got.Dropped[0] != "--ui" {
		t.Errorf("quiet flags are listed: %v", got.Dropped)
	}
}

func TestRewriteScriptReportsWhatItCannotTranslate(t *testing.T) {
	names := map[string]string{"web": "web"}
	cases := map[string]string{
		"turbo run build --filter=web...":    "dependency graph",
		"turbo run build --filter=./apps/*":  "path",
		"turbo run build --dry":              "smith explain",
		"turbo run build --parallel":         "dependencies",
		"turbo run test -- --watch":          "does not pass arguments",
		"turbo run build --filter=missing":   "names no package",
		"turbo watch dev":                    "smith dev",
		"turbo prune web --docker":           "smith run prune",
		"turbo login":                        "no smith equivalent",
		"pnpm add zod":                       "smith install -w",
		"pnpm -r exec foo":                   "not a script",
		"pnpm --filter web run test --watch": "does not pass arguments",
		"pnpm build | tee log":               "shell construct",
		"cross-env A=1 turbo run build":      "cross-env",
		"yarn workspaces foreach run build":  "smith run <task>",
	}
	for in, want := range cases {
		got := RewriteScript(in, names)
		if got.Problem == "" || !strings.Contains(got.Problem, want) {
			t.Errorf("%q: problem %q, want it to mention %q", in, got.Problem, want)
		}
		if got.Script != in {
			t.Errorf("%q: a script that cannot be translated stays as it is, got %q", in, got.Script)
		}
	}
}
