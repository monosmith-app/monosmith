package convert

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

// InstallPackages installs the packages the repository lists, the way its
// lockfile says (npm ci when there is one). It writes nothing to out.
func InstallPackages(ctx context.Context, root string) error {
	verb := "install"
	if _, err := os.Stat(filepath.Join(root, "package-lock.json")); err == nil {
		verb = "ci"
	}
	cmd := exec.CommandContext(ctx, "npm", verb, "--no-audit", "--no-fund")
	cmd.Dir = root
	if raw, err := cmd.CombinedOutput(); err != nil {
		tail := strings.TrimSpace(string(raw))
		if len(tail) > 600 {
			tail = tail[len(tail)-600:]
		}
		return fmt.Errorf("npm %s failed: %w: %s", verb, err, tail)
	}
	return nil
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
	if !c.Yes {
		if !c.Terminal {
			return problem.New("smith needs to read your projects before it can move them, and Nx describes them, so its packages must be installed first",
				"run this again with --yes to install them first",
				"or install them yourself (npm ci), then run this again")
		}
		fmt.Fprintf(c.Out, "%s [Y/n] ", InstallAsk)
		answer, _ := bufio.NewReader(c.In).ReadString('\n')
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "", "y", "yes":
		default:
			return problem.New("the packages are not installed, so smith cannot read your projects yet", "npm ci, then run this again")
		}
	}
	install := c.Install
	if install == nil {
		install = InstallPackages
	}
	fmt.Fprintln(c.Out, "Installing the packages, so smith can read your projects…")
	start := time.Now()
	if err := install(ctx, root); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "Packages installed in %s.\n\n", time.Since(start).Round(time.Second))
	return nil
}
