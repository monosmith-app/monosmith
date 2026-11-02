package convert

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEveryInstallTheGraphReadRunsKeepsTheScriptsOfTheRepositoryOff(t *testing.T) {
	for _, tc := range []struct {
		files   []string
		command string
	}{
		{nil, "npm install --no-audit --no-fund --ignore-scripts"},
		{[]string{"package-lock.json"}, "npm ci --no-audit --no-fund --ignore-scripts"},
		{[]string{"pnpm-lock.yaml"}, "pnpm install --frozen-lockfile --ignore-scripts"},
		{[]string{"yarn.lock"}, "yarn install --frozen-lockfile --ignore-scripts"},
		{[]string{"yarn.lock", ".yarnrc.yml"}, "YARN_ENABLE_SCRIPTS=false yarn install --immutable"},
	} {
		dir := t.TempDir()
		for _, name := range tc.files {
			if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if got := RepositoryTool(dir).Command(); got != tc.command {
			t.Errorf("%v: command %q, want %q", tc.files, got, tc.command)
		}
	}
}

func TestAPlanNeverInstallsAndNeverAsksEvenWithYes(t *testing.T) {
	ran := 0
	install := func(context.Context, string) error { ran++; return nil }
	for _, yes := range []bool{false, true} {
		var out bytes.Buffer
		err := EnsureInstalled(context.Background(), nxFolder(t, false), InstallChoice{Yes: yes, DryRun: true, Terminal: true, In: strings.NewReader("y\n"), Out: &out, Install: install})
		if err == nil || ran != 0 {
			t.Fatalf("yes=%v: err %v, installs %d", yes, err, ran)
		}
		for _, want := range []string{"nothing was installed", "smith would run `npm install --no-audit --no-fund --ignore-scripts`", "it runs no script of the repository"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("yes=%v: the message lacks %q: %v", yes, want, err)
			}
		}
		if out.Len() != 0 {
			t.Errorf("a plan asked or told something on its screen: %q", out.String())
		}
	}
}

func TestAPlanOfAReadyWorkspaceNeedsNoInstall(t *testing.T) {
	if err := EnsureInstalled(context.Background(), nxFolder(t, true), InstallChoice{DryRun: true, Out: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}
}

func TestTheQuestionAndTheScreenNameTheExactCommand(t *testing.T) {
	var out bytes.Buffer
	err := EnsureInstalled(context.Background(), nxFolder(t, false), InstallChoice{Terminal: true, In: strings.NewReader("y\n"), Out: &out, Install: func(context.Context, string) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), "npm install --no-audit --no-fund --ignore-scripts") < 2 {
		t.Errorf("the question and the progress line must both name the command:\n%s", out.String())
	}
}
