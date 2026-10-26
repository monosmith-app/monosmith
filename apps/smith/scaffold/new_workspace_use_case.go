package scaffold

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// NewOptions configure a new workspace.
type NewOptions struct {
	Name string
	// Dir is where to create it; default ./<name>. It may already exist and
	// hold files smith does not write (a fresh clone with a LICENSE).
	Dir          string
	Scope        string
	ModulePrefix string
	Providers    []string
	// Recipe is written as [workspace] recipe: the recipe `smith add`
	// creates projects from.
	Recipe  string
	Version string
	Out     io.Writer
	// NoCommit leaves the files uncommitted (smith adopt commits once, after
	// generating the pipelines for the projects it found).
	NoCommit bool
}

// NewWorkspace creates a workspace: smith.toml, the root files, a git
// repository, the pipelines, and a first commit when git has an identity.
func NewWorkspace(opts NewOptions) (string, error) {
	if err := ValidateName("workspace", opts.Name); err != nil {
		return "", err
	}
	dir := opts.Dir
	if dir == "" {
		dir = opts.Name
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(dir, workspace.FileName)); err == nil {
		return "", fmt.Errorf("%s already holds a smith workspace", dir)
	}
	if err := checkRecipe(dir, opts.Recipe, opts.Out); err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	git := gitclient.New(dir)
	// A repository smith creates is the workspace's alone, so its first commit
	// takes everything in it. In someone else's repository smith commits only
	// the files it wrote, and below the top of one (a folder of a bigger
	// repository) it commits nothing: the user decides what goes in.
	created, nested := false, false
	if !git.IsRepository() {
		if err := git.Init(); err != nil {
			return "", err
		}
		created = true
	} else if prefix, err := git.Prefix(); err == nil && prefix != "" {
		nested = true
	}
	if opts.Scope == "" {
		opts.Scope = "@" + opts.Name
	}
	if !strings.HasPrefix(opts.Scope, "@") {
		opts.Scope = "@" + opts.Scope
	}
	if opts.ModulePrefix == "" {
		opts.ModulePrefix = modulePrefixFromRemote(git, opts.Name)
	}
	if len(opts.Providers) == 0 {
		opts.Providers = []string{"github"}
	}
	quoted := make([]string, len(opts.Providers))
	for i, p := range opts.Providers {
		quoted[i] = fmt.Sprintf("%q", p)
	}
	config := fmt.Sprintf(configTemplate, opts.Name, opts.Scope, opts.ModulePrefix, strings.Join(quoted, ", "))
	if opts.Recipe != "" {
		config = withRecipe(config, opts.ModulePrefix, opts.Recipe)
	}
	cfg, err := workspace.Parse([]byte(config))
	if err != nil {
		return "", err
	}
	files := map[string]string{
		workspace.FileName: config,
		".gitignore":       gitignore,
		".gitattributes":   gitattributes,
		"README.md":        fmt.Sprintf(readme, opts.Name),
	}
	var written []string
	for name, body := range files {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil && name != workspace.FileName {
			fmt.Fprintf(opts.Out, "  kept your %s\n", name)
			if name == ".gitignore" {
				changed, err := ignoreCache(p, opts.Out)
				if err != nil {
					return "", err
				}
				if changed {
					written = append(written, name)
				}
			}
			continue
		} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			return "", err
		}
		written = append(written, name)
		fmt.Fprintf(opts.Out, "  wrote %s\n", name)
	}
	pipelines, _, err := cipipeline.Write(dir, cfg, cipipeline.NewModel(cipipeline.Contributions{}, opts.Version, cfg.Release.Branch))
	if err != nil {
		return "", err
	}
	for _, f := range pipelines {
		fmt.Fprintf(opts.Out, "  wrote %s\n", f)
	}
	written = append(written, pipelines...)
	if opts.NoCommit {
		return dir, nil
	}
	if nested {
		top, _ := git.TopLevel()
		fmt.Fprintf(opts.Out, "  not committed: %s is inside the git repository at %s; commit the workspace there when ready\n", dir, filepath.FromSlash(top))
		return dir, nil
	}
	message := "chore: create the " + opts.Name + " workspace with smith"
	commit := func() error { return git.CommitPaths(message, written) }
	if created {
		commit = func() error { return git.CommitAll(message) }
	}
	if err := commit(); err != nil {
		fmt.Fprintf(opts.Out, "  not committed (%v); commit when ready\n", firstLine(err.Error()))
	} else {
		fmt.Fprintln(opts.Out, "  committed")
	}
	return dir, nil
}

// ignoreCache adds .smith/ (the task cache) to a .gitignore smith kept, unless
// it already ignores it, and says whether it changed the file.
func ignoreCache(path string, out io.Writer) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if l := strings.TrimSpace(line); l == ".smith/" || l == ".smith" || l == "/.smith/" || l == "/.smith" {
			return false, nil
		}
	}
	text := string(raw)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	if err := os.WriteFile(path, []byte(text+"\n# smith\n.smith/\n"), 0o644); err != nil {
		return false, err
	}
	fmt.Fprintln(out, "  added .smith/ to your .gitignore")
	return true, nil
}

// modulePrefixFromRemote derives github.com/<owner>/<repo> from the origin
// remote, falling back to example.com/<name>.
func modulePrefixFromRemote(git *gitclient.Git, name string) string {
	if prefix := remoteModulePrefix(git); prefix != "" {
		return prefix
	}
	return "example.com/" + name
}

// remoteModulePrefix is github.com/<owner>/<repo> from the origin remote, or
// "" when the repository has none.
func remoteModulePrefix(git *gitclient.Git) string {
	url, err := git.RemoteURL("origin")
	if err != nil || strings.TrimSpace(url) == "" {
		return ""
	}
	url = strings.TrimSuffix(strings.TrimSpace(url), ".git")
	for _, prefix := range []string{"https://", "http://", "ssh://", "git@"} {
		url = strings.TrimPrefix(url, prefix)
	}
	url = strings.Replace(url, ":", "/", 1)
	if i := strings.Index(url, "@"); i >= 0 {
		url = url[i+1:]
	}
	if url == "" {
		return ""
	}
	return strings.ToLower(url[:1]) + url[1:]
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}
