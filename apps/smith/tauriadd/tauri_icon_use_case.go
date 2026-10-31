package tauriadd

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
)

// iconCommand is what makes the icons: the project's own Tauri CLI, given the
// logo by its absolute path because it runs inside the project.
func iconCommand(logo string) []string {
	return []string{"npx", "--no-install", "tauri", "icon", logo}
}

// tauriProject is the Tauri app to work on: the named one, or the only one.
func tauriProject(s *session.Session, name string) (root, projectName string, err error) {
	var found []string
	roots := map[string]string{}
	for _, p := range s.Graph.Projects {
		if name != "" && p.Name != name {
			continue
		}
		if tauriapp.Detect(s.Root, p.Root) {
			found = append(found, p.Name)
			roots[p.Name] = p.Root
		}
	}
	switch {
	case len(found) == 1:
		return roots[found[0]], found[0], nil
	case name != "":
		return "", "", problem.New(name+" has no src-tauri", "smith tauri add "+name)
	case len(found) == 0:
		return "", "", problem.New("no project has a src-tauri", "smith tauri add <project>")
	}
	return "", "", problem.New("several projects have a src-tauri", "name one with -p: "+found[0]+", "+found[1])
}

// MakeIcons runs the Tauri CLI's icon command on logo for the project, which
// overwrites its src-tauri/icons.
func MakeIcons(s *session.Session, projectName, logo string) (string, error) {
	abs, err := filepath.Abs(logo)
	if err != nil {
		return "", err
	}
	if st, statErr := os.Stat(abs); statErr != nil || st.IsDir() {
		return "", problem.New(logo+" is not a file", "give the path of a square PNG, 1024x1024 is best")
	}
	root, name, err := tauriProject(s, projectName)
	if err != nil {
		return "", err
	}
	args := iconCommand(abs)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = filepath.Join(s.Root, filepath.FromSlash(root))
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
		return name, problem.New("the Tauri CLI could not make the icons: "+err.Error(), "is @tauri-apps/cli installed? smith install, or smith tauri add "+name)
	}
	return name, nil
}
