package adoption

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	if install {
		cmd := exec.Command("npm", "install")
		cmd.Dir, cmd.Stdout, cmd.Stderr = r.Root, out, out
		fmt.Fprintln(out, "  npm install ...")
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(out, "  not done: npm install failed (%v); run it yourself\n", err)
		}
	}
	return nil
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
