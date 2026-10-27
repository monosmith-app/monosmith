package adoption

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/lockfilesync"
)

// CompleteWorkspace does the mechanical steps smith doctor would otherwise
// ask for right after adoption: go.work for the Go projects, the root
// package.json "workspaces" for the Node ones, node_modules in .gitignore,
// and, when install is set, npm install. What it cannot do safely it says,
// and leaves to doctor's own remedy.
func CompleteWorkspace(r Report, install, eslint bool, out io.Writer) error {
	var goRoots, nodeRoots []string
	for _, p := range r.Projects {
		switch p.Language {
		case "go":
			goRoots = append(goRoots, p.Root)
		case "node":
			nodeRoots = append(nodeRoots, p.Root)
		}
	}
	if len(goRoots) > 0 {
		if err := writeGoWork(r.Root, goRoots, out); err != nil {
			return err
		}
	}
	if len(nodeRoots) == 0 {
		if eslint {
			fmt.Fprintln(out, "  --eslint skipped: there are no Node projects")
		}
		return nil
	}
	manifest := filepath.Join(r.Root, "package.json")
	raw, _ := os.ReadFile(manifest)
	updated, changed, err := withWorkspaces(raw, r.Name, workspacePatterns(nodeRoots))
	if err != nil {
		fmt.Fprintf(out, "  not done: %v; add the \"workspaces\" by hand\n", err)
	} else if changed {
		if err := os.WriteFile(manifest, updated, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(out, "  wrote \"workspaces\" in package.json: %s\n", strings.Join(workspacePatterns(nodeRoots), ", "))
	}
	if err := ignoreNodeModules(r.Root, out); err != nil {
		return err
	}
	if eslint {
		if err := AddESLintConfig(r.Root, out); err != nil {
			return err
		}
	}
	settleLockfile(lockfilesync.Options{Root: r.Root, Out: out, Create: install, LockOnly: !install}, install || changed, out)
	return nil
}

// settleLockfile leaves package-lock.json in step with the package.json that
// adoption just edited, when run is true: a full npm install when the person
// asked for one (--install), otherwise only the lockfile, and only when
// there is one. A failure is told, not returned: the adoption itself is done.
func settleLockfile(o lockfilesync.Options, run bool, out io.Writer) {
	if !run {
		return
	}
	if o.Create {
		fmt.Fprintln(out, "  npm install ...")
	}
	res, err := lockfilesync.Sync(o)
	switch {
	case err != nil:
		fmt.Fprintf(out, "  not done: package-lock.json is not in step with package.json (%v)\n", err)
	case res.Skipped != "":
		return
	case o.Create:
		fmt.Fprintln(out, "  npm install done")
	default:
		fmt.Fprintln(out, "  package-lock.json updated for the new \"workspaces\"")
	}
	for _, note := range res.Notes {
		fmt.Fprintf(out, "    note: %s\n", note)
	}
}

// writeGoWork runs go work init and go work use for the Go projects, unless
// the repository has a go.work already.
func writeGoWork(root string, goRoots []string, out io.Writer) error {
	if _, err := os.Stat(filepath.Join(root, "go.work")); err == nil {
		return nil
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		fmt.Fprintln(out, "  not done: go is not installed, so go.work was not written; run go work init later")
		return nil
	}
	dirs := make([]string, len(goRoots))
	for i, d := range goRoots {
		dirs[i] = "./" + d
	}
	for _, args := range [][]string{{"work", "init"}, append([]string{"work", "use"}, dirs...)} {
		cmd := exec.Command(goBin, args...)
		cmd.Dir = root
		if msg, err := cmd.CombinedOutput(); err != nil {
			fmt.Fprintf(out, "  not done: go %s failed (%s); run it yourself\n", strings.Join(args[:2], " "), strings.TrimSpace(string(msg)))
			return nil
		}
	}
	fmt.Fprintf(out, "  wrote go.work: %s\n", strings.Join(dirs, ", "))
	return nil
}

// ignoreNodeModules adds node_modules/ to .gitignore when nothing ignores it,
// so an install never lands in the adoption commit.
func ignoreNodeModules(root string, out io.Writer) error {
	path := filepath.Join(root, ".gitignore")
	body, _ := os.ReadFile(path)
	for _, line := range strings.Split(string(body), "\n") {
		switch strings.TrimSpace(line) {
		case "node_modules", "node_modules/", "/node_modules", "/node_modules/", "**/node_modules", "**/node_modules/":
			return nil
		}
	}
	text := string(body)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	if err := os.WriteFile(path, []byte(text+"node_modules/\n"), 0o644); err != nil {
		return err
	}
	fmt.Fprintln(out, "  added node_modules/ to your .gitignore")
	return nil
}
