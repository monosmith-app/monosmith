package convertplan

import (
	"strings"
	"testing"
)

func TestSplitCommandLine(t *testing.T) {
	for line, want := range map[string]string{
		`go build -o ../../dist/x .`:                     `go|build|-o|../../dist/x|.`,
		`node tools/go-slice-check.cjs apps/smith`:       `node|tools/go-slice-check.cjs|apps/smith`,
		`node -e "console.log('a b')"`:                   `node|-e|console.log('a b')`,
		`eslint . --ignore-pattern "apps/**"`:            `eslint|.|--ignore-pattern|apps/**`,
		`go build "-gcflags=all=-N -l" -o out .`:         `go|build|-gcflags=all=-N -l|-o|out|.`,
		`echo 'it''s'`:                                   `echo|its`,
		`printf "say \"hi\""`:                            `printf|say "hi"`,
		`python3 -m pytest`:                              `python3|-m|pytest`,
		"  go   vet  ./...  ":                            `go|vet|./...`,
		`a\ b`:                                           `a b`,
		`node -e "require('x')({\"k\":1})"`:              `node|-e|require('x')({"k":1})`,
		`CGO_ENABLED=0 GOOS=linux go build .`:            `go|build|.`,
		`go run github.com/air-verse/air@v1.61.7 --x y`:  `go|run|github.com/air-verse/air@v1.61.7|--x|y`,
		`npx tsc -p tsconfig.lib.json --noEmit`:          `npx|tsc|-p|tsconfig.lib.json|--noEmit`,
		`node -e "const a=[1,2];process.exit(a.length)"`: `node|-e|const a=[1,2];process.exit(a.length)`,
	} {
		got, err := SplitCommandLine(line)
		if err != nil {
			t.Errorf("%s: %v", line, err)
			continue
		}
		if strings.Join(got.Argv, "|") != want {
			t.Errorf("%s: got %q want %q", line, strings.Join(got.Argv, "|"), want)
		}
	}
	env, _ := SplitCommandLine(`CGO_ENABLED=0 GOOS=linux go build .`)
	if env.Env["CGO_ENABLED"] != "0" || env.Env["GOOS"] != "linux" {
		t.Fatalf("env %v", env.Env)
	}
}

func TestSplitCommandLineRefusesShellOnly(t *testing.T) {
	for line, want := range map[string]string{
		`npm run a && npm run b`:      `"&&"`,
		`cat a | grep b`:              `"|"`,
		`echo x > out.txt`:            `">"`,
		`go build -X main.v=$VERSION`: `"$"`,
		`echo "$HOME"`:                `"$"`,
		`rm -rf dist/*`:               `glob`,
		`echo $(date)`:                `"$"`,
		`a; b`:                        `";"`,
		`echo 'open`:                  `unterminated`,
		``:                            `no command`,
	} {
		_, err := SplitCommandLine(line)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: want error containing %s, got %v", line, want, err)
		}
	}
}
