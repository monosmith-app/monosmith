package convert

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/convertplan"
)

func TestPruneLockfileDropsNxAtEveryDepth(t *testing.T) {
	lock := `{
  "name": "ws",
  "lockfileVersion": 3,
  "packages": {
    "": { "name": "ws" },
    "node_modules/@mnci/eslint-config": { "version": "0.3.31" },
    "node_modules/@nx/eslint-plugin": { "version": "23.2.0", "optional": true },
    "node_modules/nx": { "version": "23.2.0" },
    "node_modules/@nx-go/nx-go/node_modules/@nx/devkit": { "version": "22.7.12" },
    "node_modules/nxtend": { "version": "1.0.0" },
    "node_modules/eslint": { "version": "10.8.1" }
  }
}
`
	out, dropped, err := PruneLockfile([]byte(lock), convertplan.NxPackage)
	if err != nil {
		t.Fatal(err)
	}
	if len(dropped) != 3 {
		t.Fatalf("dropped %v, want the three Nx entries", dropped)
	}
	text := string(out)
	for _, gone := range []string{`"node_modules/nx"`, "@nx/eslint-plugin", "@nx/devkit"} {
		if strings.Contains(text, gone) {
			t.Errorf("%s is still in the lockfile:\n%s", gone, text)
		}
	}
	for _, kept := range []string{"@mnci/eslint-config", "nxtend", `"node_modules/eslint"`, `"lockfileVersion": 3`} {
		if !strings.Contains(text, kept) {
			t.Errorf("%s was removed:\n%s", kept, text)
		}
	}
}

func TestPruneLockfileWithoutNxIsUnchanged(t *testing.T) {
	lock := "{\n\t\"packages\": {\"node_modules/eslint\": {}}\n}\n"
	out, dropped, err := PruneLockfile([]byte(lock), convertplan.NxPackage)
	if err != nil || len(dropped) != 0 || string(out) != lock {
		t.Fatalf("got %q %v %v", out, dropped, err)
	}
}
