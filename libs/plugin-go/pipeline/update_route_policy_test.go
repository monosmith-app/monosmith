package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

const goBuildInfo = "golangci-lint: go1.26\n\tpath\tgithub.com/golangci/golangci-lint/v2/cmd/golangci-lint\n"

// goMachine is fakeMachine plus build info for the files in buildInfo.
func goMachine(t *testing.T, onPath []string, says map[string]string, bin string, buildInfo map[string]string) {
	t.Helper()
	fakeMachine(t, onPath, says, bin)
	inner := probe
	probe = func(dir string, argv []string) (string, error) {
		if len(argv) > 2 && argv[0] == "go" && argv[1] == "version" {
			if out, ok := buildInfo[argv[3]]; ok {
				return out, nil
			}
			return "", os.ErrNotExist
		}
		return inner(dir, argv)
	}
}

func TestGoBinCopyIsUpdatedWithGoInstallIntoItsOwnFolder(t *testing.T) {
	root := workspaceRoot(t)
	bin, file := putLint(t, "")
	goMachine(t, []string{file}, map[string]string{file: "golangci-lint has version 2.13.0 built with go1.26"}, bin, map[string]string{file: goBuildInfo})
	f := lintFinding(root, "2.15.0")
	options := f.Fix.Choices[0].Options
	if len(options) != 3 || !strings.Contains(options[0].Label, "recommended") {
		t.Fatalf("three answers, the first stays the recommended one: %+v", options)
	}
	update := options[2]
	step := update.Fix.Steps[0]
	if update.Disabled != "" || step.Env[0] != "GOBIN="+bin || step.Argv[0] != "go" || step.Argv[1] != "install" || !strings.HasSuffix(step.Argv[2], "@v2.15.0") {
		t.Fatalf("go install of the pin, writing into the folder of the copy found: %+v", update)
	}
	for _, want := range []string{"Update " + file + " to 2.15.0", "every project that uses this copy gets it", "GOBIN=" + bin + " go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0"} {
		if !strings.Contains(update.Label, want) {
			t.Errorf("the label states the command and the reach, and lacks %q: %s", want, update.Label)
		}
	}
	if len(update.Fix.Files) != 0 || len(options[0].Fix.Steps) != 1 || options[0].Fix.Steps[0].Env[0] != "GOBIN="+toolrun.ManagedBinDir(root) {
		t.Errorf("the update writes no file and the managed install is unchanged")
	}
}

func TestACopyOutsideEveryGoBinFolderIsGreyedWithAReason(t *testing.T) {
	root := workspaceRoot(t)
	bin, _ := putLint(t, "")
	elsewhere := filepath.Join(t.TempDir(), "npm", "golangci-lint")
	goMachine(t, []string{elsewhere}, map[string]string{elsewhere: "golangci-lint has version 2.13.0 built with go1.26"}, bin, map[string]string{elsewhere: goBuildInfo})
	update := lintFinding(root, "2.15.0").Fix.Choices[0].Options[2]
	if update.Disabled != "smith does not know how this copy was installed" || len(update.Fix.Steps) != 0 {
		t.Fatalf("an unknown origin is not guessed: %+v", update)
	}
}

func TestACopyInAGoBinFolderWithoutGoBuildInfoIsGreyed(t *testing.T) {
	root := workspaceRoot(t)
	bin, file := putLint(t, "")
	goMachine(t, []string{file}, map[string]string{file: "golangci-lint has version 2.13.0 built with go1.26"}, bin, nil)
	if update := lintFinding(root, "2.15.0").Fix.Choices[0].Options[2]; update.Disabled == "" {
		t.Fatalf("a program go cannot read is not updated with go install: %+v", update)
	}
}

func TestGoBinFoldersListGobinGopathAndHome(t *testing.T) {
	fakeMachine(t, nil, nil, "/a/bin")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home folder")
	}
	got := goBinFolders(t.TempDir())
	if len(got) != 3 || got[0] != "/a/bin" || !sameFile(got[1], filepath.Join(string(filepath.Separator)+"a", "bin")) || got[2] != filepath.Join(home, "go", "bin") {
		t.Fatalf("GOBIN, each GOPATH entry's bin, then ~/go/bin: %v", got)
	}
}

func TestACopyInSmithsOwnFolderGetsNoUpdateAnswer(t *testing.T) {
	root := workspaceRoot(t)
	_, managed := putLint(t, toolrun.ManagedBinDir(root))
	goMachine(t, nil, map[string]string{managed: "golangci-lint has version 2.13.0 built with go1.26"}, "", nil)
	if options := lintFinding(root, "2.15.0").Fix.Choices[0].Options; len(options) != 2 {
		t.Fatalf("the first answer already replaces smith's own copy: %+v", options)
	}
}

// workspaceRoot is a workspace folder with a smith.toml, which the keep answer needs.
func workspaceRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "smith.toml"), []byte("[workspace]\nname = \"w\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}
