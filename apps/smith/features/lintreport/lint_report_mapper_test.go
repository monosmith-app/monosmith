package lintreport

import (
	"strings"
	"testing"
)

const eslintOutput = `mitt:deps-check ▶ {smith} npm-deps-check packages/mitt
mitt:deps-check ✓ 40ms
mitt:lint ▶ npm run lint
mitt:lint │
mitt:lint │ > eslint .
mitt:lint │ C:\ws\packages\mitt\index.d.ts
mitt:lint │    1:48   error  Extra semicolon                                       @stylistic/semi
mitt:lint │    2:62   error  Extra semicolon                                       @stylistic/semi
mitt:lint │    5:21   error  Prefer ` + "`A`" + ` over ` + "`B`" + `  unicorn/consistent-compound-words
mitt:lint │    8:1    error  Expected indentation of 2 spaces but found 4          @stylistic/indent
mitt:lint │    9:1    warning  Unexpected console statement                        no-console
mitt:lint │
mitt:lint │ ✖ 835 problems (832 errors, 3 warnings)
mitt:lint │   502 errors and 1 warning potentially fixable with the ` + "`--fix`" + ` option.
mitt:lint │ npm error command failed
mitt:lint × exited with code 1
  × mitt:lint exited with code 1
× 1/2 tasks succeeded (0 from cache) in 10.779s
`

func TestSummarizeCountsESLintProblemsAndNamesTheFix(t *testing.T) {
	got := Summarize("mitt", eslintOutput)
	if len(got) != 1 {
		t.Fatalf("items %+v", got)
	}
	if want := "the lint of mitt finds 835 problem(s) in its code, most from @stylistic/semi (2), @stylistic/indent (1), no-console (1)"; got[0].Problem != want {
		t.Errorf("problem\n%s\nwant\n%s", got[0].Problem, want)
	}
	if !strings.HasPrefix(got[0].Remedy, "smith lint --fix -p mitt repairs 503 of them") {
		t.Errorf("remedy %q", got[0].Remedy)
	}
}

const ruffOutput = `file-analyzer:slice-check ✓ 38ms
file-analyzer:lint ▶ python -m ruff check .
file-analyzer:lint │ I001 [*] Import block is un-sorted or un-formatted
file-analyzer:lint │  --> file_analyzer.py:1:1
file-analyzer:lint │ F401 [*] ` + "`os`" + ` imported but unused
file-analyzer:lint │ main.py:3:1: E722 Do not use bare ` + "`except`" + `
file-analyzer:lint │ main.py:9:1: E722 Do not use bare ` + "`except`" + `
file-analyzer:lint │ Found 5 errors.
file-analyzer:lint │ [*] 2 fixable with ` + "`--fix`" + `.
file-analyzer:lint × exited with code 1
`

func TestSummarizeCountsRuffProblemsInBothFormats(t *testing.T) {
	got := Summarize("file-analyzer", ruffOutput)
	if len(got) != 1 || !strings.Contains(got[0].Problem, "finds 5 problem(s)") || !strings.Contains(got[0].Problem, "E722 (2), F401 (1), I001 (1)") {
		t.Fatalf("items %+v", got)
	}
	if !strings.Contains(got[0].Remedy, "smith lint --fix -p file-analyzer repairs 2 of them") {
		t.Errorf("remedy %q", got[0].Remedy)
	}
}

func TestSummarizeWithoutFixableProblemsListsThem(t *testing.T) {
	got := Summarize("a", "a:lint │ Found 1 error.\na:lint │ x.py:1:1: E722 Do not\na:lint × exited with code 1\n")
	if len(got) != 1 || strings.Contains(got[0].Remedy, "--fix") || !strings.HasPrefix(got[0].Remedy, "smith lint -p a lists them") {
		t.Errorf("items %+v", got)
	}
}

const sliceOutput = `sampleproject:slice-check ▶ {smith} slices check --lang python .
sampleproject:deps-check ▶ {smith} py-deps-check packages/sampleproject
sampleproject:slice-check │ src/sample/simple.py - only __init__.py and __main__.py sit at the top of a package; move this into a slice package
sampleproject:slice-check │ smith: 1 file(s) or folder(s) break the vertical-slice layout; an exception is written in the project README
sampleproject:slice-check × exited with code 1
sampleproject:deps-check │ noxfile.py imports nox, which pyproject.toml does not declare
sampleproject:deps-check │ smith: packages/sampleproject imports 1 module(s) its pyproject.toml does not declare
sampleproject:deps-check × exited with code 1
  × sampleproject:deps-check exited with code 1
  » sampleproject:lint because sampleproject:deps-check did not succeed
  × sampleproject:slice-check exited with code 1
`

func TestSummarizeNamesTheSliceCheckWithItsWaysOnAndTheTaskThatStoppedTheLint(t *testing.T) {
	got := Summarize("sampleproject", sliceOutput)
	if len(got) != 2 {
		t.Fatalf("items %+v", got)
	}
	slice := got[0]
	if !strings.Contains(slice.Problem, "1 file(s) or folder(s) break the layout") || !strings.Contains(slice.Problem, "first: src/sample/simple.py") {
		t.Errorf("slice problem %q", slice.Problem)
	}
	for _, want := range []string{"smith run slice-check -p sampleproject", "README", `exclude under [conventions.vertical-slices]`} {
		if !strings.Contains(slice.Remedy, want) {
			t.Errorf("slice remedy lacks %q: %q", want, slice.Remedy)
		}
	}
	if want := "the lint of sampleproject did not run: its task deps-check fails (packages/sampleproject imports 1 module(s) its pyproject.toml does not declare)"; got[1].Problem != want {
		t.Errorf("deps problem %q", got[1].Problem)
	}
}

func TestSummarizeOfAnOutputWithoutAFailedTaskNamesTheCommand(t *testing.T) {
	got := Summarize("a", "no project matches a\n")
	if len(got) != 1 || got[0].Remedy != "smith run lint -p a shows why" {
		t.Errorf("items %+v", got)
	}
}
