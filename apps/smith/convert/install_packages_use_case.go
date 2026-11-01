package convert

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// InstallAsk is the question before smith installs an Nx workspace's packages.
const InstallAsk = "smith needs to read your projects before it can move them, and Nx describes them, so the packages have to be installed first (this can take a couple of minutes). Install them now?"

// NxNeedsInstall is true when the folder is an Nx workspace whose packages are
// not installed, so Nx cannot describe its projects yet.
func NxNeedsInstall(root string) bool {
	has := func(rel string) bool {
		_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
		return err == nil
	}
	return has("nx.json") && !has("turbo.json") && !has("node_modules/nx")
}

// PackageTool is the package manager a repository installs with.
type PackageTool struct {
	// Name is the program: npm, pnpm or yarn.
	Name string
	// Args are the arguments of a clean install with the lockfile.
	Args []string
}

// Command is the install as a person types it: "pnpm install --frozen-lockfile".
func (t PackageTool) Command() string { return t.Name + " " + strings.Join(t.Args, " ") }

// RepositoryTool picks the package manager the repository's lockfile names:
// pnpm for pnpm-lock.yaml, yarn for yarn.lock, npm otherwise (npm ci when
// there is a package-lock.json). A pnpm or yarn workspace has to be installed
// with its own tool: npm cannot read its catalog: and workspace: ranges.
func RepositoryTool(root string) PackageTool {
	has := func(rel string) bool {
		_, err := os.Stat(filepath.Join(root, rel))
		return err == nil
	}
	switch {
	case has("pnpm-lock.yaml") || has("pnpm-workspace.yaml"):
		return PackageTool{Name: "pnpm", Args: []string{"install", "--frozen-lockfile"}}
	case has("yarn.lock") && has(".yarnrc.yml"):
		return PackageTool{Name: "yarn", Args: []string{"install", "--immutable"}}
	case has("yarn.lock"):
		return PackageTool{Name: "yarn", Args: []string{"install", "--frozen-lockfile"}}
	case has("package-lock.json"):
		return PackageTool{Name: "npm", Args: []string{"ci", "--no-audit", "--no-fund"}}
	}
	return PackageTool{Name: "npm", Args: []string{"install", "--no-audit", "--no-fund"}}
}

// InstallPackages installs the packages the repository lists with the
// repository's own package manager (RepositoryTool). It writes nothing to out.
func InstallPackages(ctx context.Context, root string) error {
	tool := RepositoryTool(root)
	cmd := exec.CommandContext(ctx, tool.Name, tool.Args...)
	cmd.Dir = root
	if raw, err := cmd.CombinedOutput(); err != nil {
		tail := strings.TrimSpace(string(raw))
		if len(tail) > 600 {
			tail = tail[len(tail)-600:]
		}
		return fmt.Errorf("%s failed: %w: %s", tool.Command(), err, tail)
	}
	return nil
}

// installLeftovers are the paths an Nx install leaves in the repository that
// git lists as new, in the order they are named to the person.
var installLeftovers = []string{".nx", "package-lock.json"}

// existingLeftovers lists which of installLeftovers exist in root.
func existingLeftovers(root string) []string {
	var found []string
	for _, name := range installLeftovers {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			found = append(found, name)
		}
	}
	return found
}

// InstallChoice says how the command line gets the packages installed.
type InstallChoice struct {
	// Yes installs without asking.
	Yes bool
	// Terminal is true when someone can answer a question.
	Terminal bool
	// In and Out are where the question is asked and the progress is told.
	In  io.Reader
	Out io.Writer
	// Install does the install; InstallPackages when nil.
	Install func(ctx context.Context, root string) error
}

// EnsureInstalled installs an Nx workspace's packages when they are missing:
// without asking under --yes, after a question in a terminal, and otherwise it
// stops and says what to do. It does nothing for a workspace that is ready.
func EnsureInstalled(ctx context.Context, root string, c InstallChoice) error {
	if !NxNeedsInstall(root) {
		return nil
	}
	tool := RepositoryTool(root)
	if !c.Yes {
		if !c.Terminal {
			if tool.Name != "npm" {
				return problem.New("this is a "+tool.Name+" workspace: smith needs Nx's packages installed to read your projects, and only "+tool.Name+" can install them here",
					tool.Name+" install, then run this again",
					"or run this again with --yes: smith then installs them with "+tool.Name)
			}
			return problem.New("smith needs to read your projects before it can move them, and Nx describes them, so its packages must be installed first",
				"run this again with --yes to install them first",
				"or install them yourself ("+tool.Command()+"), then run this again")
		}
		question := InstallAsk
		if tool.Name != "npm" {
			question = strings.Replace(question, "Install them now?", "This is a "+tool.Name+" workspace, so smith installs them with "+tool.Name+". Install them now?", 1)
		}
		fmt.Fprintf(c.Out, "%s [Y/n] ", question)
		answer, _ := bufio.NewReader(c.In).ReadString('\n')
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "", "y", "yes":
		default:
			return problem.New("the packages are not installed, so smith cannot read your projects yet", tool.Command()+", then run this again")
		}
	}
	install := c.Install
	if install == nil {
		install = InstallPackages
	}
	fmt.Fprintf(c.Out, "Installing the packages with %s, so smith can read your projects…\n", tool.Name)
	before := existingLeftovers(root)
	start := time.Now()
	if err := install(ctx, root); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "Packages installed in %s.\n", time.Since(start).Round(time.Second))
	var created []string
	for _, name := range existingLeftovers(root) {
		if !slices.Contains(before, name) {
			created = append(created, name)
		}
	}
	if len(created) > 0 {
		fmt.Fprintf(c.Out, "The install created %s, which git lists as new. A later smith convert ignores them; delete them if you do not want them.\n", strings.Join(created, " and "))
	}
	fmt.Fprintln(c.Out)
	return nil
}
