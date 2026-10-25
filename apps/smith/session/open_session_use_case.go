package session

import (
	"fmt"

	"errors"
	"github.com/monosmith-app/monosmith/apps/smith/remotecache"
	"io"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Session is an open workspace.
type Session struct {
	Root    string
	Config  workspace.Config
	Git     *gitclient.Git
	Plugins *pluginhost.PluginSet
	// Stacks are the stacks the workspace can use.
	Stacks *pluginhost.Catalog
	Graph  *projectgraph.Graph
	// Files are tracked files plus untracked files that are not ignored.
	Files   []string
	Version string
	// allowCycles keeps the graph open when projects import each other in a
	// cycle (Options.AllowCycles).
	allowCycles bool
}

// Options configure Open.
type Options struct {
	Dir     string
	Version string
	// Exact looks for smith.toml from Dir upwards only, ignoring
	// SMITH_WORKSPACE_ROOT: for a directory that is not the workspace the
	// process runs in, such as the scratch copy of a conversion.
	Exact bool
	// Log receives plugin output; nil means stderr.
	Log io.Writer
	// Preview opens Dir itself with this configuration instead of finding
	// smith.toml: a look at a repository before it is a workspace (smith adopt).
	Preview *workspace.Config
	// AllowCycles opens the workspace although projects import each other in
	// a cycle, for the commands that report it (doctor, graph). Without it
	// Open fails with the cycle, as every command that needs an order must.
	AllowCycles bool
}

// Open finds the workspace containing opts.Dir and loads everything a command
// needs. The caller must Close it.
func Open(opts Options) (*Session, error) {
	var root string
	var cfg workspace.Config
	if opts.Preview != nil {
		dir, err := filepath.Abs(opts.Dir)
		if err != nil {
			return nil, err
		}
		root, cfg = dir, *opts.Preview
	} else {
		find := workspace.FindRoot
		if opts.Exact {
			find = workspace.FindRootFrom
		}
		found, err := find(opts.Dir)
		if err != nil {
			return nil, err
		}
		loaded, err := workspace.Load(found)
		if err != nil {
			return nil, err
		}
		root, cfg = found, loaded
	}
	git := gitclient.New(root)
	if !git.IsRepository() {
		return nil, errors.New("the workspace is not a git repository; run `git init` at " + root)
	}
	files, err := git.Files()
	if err != nil {
		return nil, err
	}
	plugins, err := pluginhost.Start(pluginhost.Options{WorkspaceRoot: root, Config: cfg, Version: opts.Version, Log: opts.Log})
	if err != nil {
		return nil, err
	}
	catalog, err := pluginhost.NewCatalog(plugins, cfg)
	if err != nil {
		plugins.Close()
		return nil, err
	}
	graph, err := discover(opts.AllowCycles, files, plugins, catalog, cfg)
	if err != nil {
		plugins.Close()
		return nil, err
	}
	return &Session{Root: root, Config: cfg, Git: git, Plugins: plugins, Stacks: catalog, Graph: graph, Files: files, Version: opts.Version, allowCycles: opts.AllowCycles}, nil
}

// discover builds the project graph, refusing a cycle unless allowed.
func discover(allowCycles bool, files []string, plugins *pluginhost.PluginSet, catalog *pluginhost.Catalog, cfg workspace.Config) (*projectgraph.Graph, error) {
	if allowCycles {
		return projectgraph.DiscoverAllowingCycles(files, plugins, catalog, cfg)
	}
	return projectgraph.Discover(files, plugins, catalog, cfg)
}

// LocalCacheDir is the directory of the workspace's local cache.
func (s *Session) LocalCacheDir() string {
	return LocalCacheDir(s.Root, s.Config, os.Getenv)
}

// Storage opens the configured cache.
func (s *Session) Storage() (taskcache.Storage, error) {
	dir := s.LocalCacheDir()
	local, err := taskcache.NewLocalStorage(dir)
	if err != nil {
		return nil, err
	}
	remote := s.Config.Cache.Remote
	if env := os.Getenv("SMITH_CACHE_REMOTE"); env != "" {
		remote = env
	}
	if remote == "" {
		return local, nil
	}
	warn := func(msg string) { fmt.Fprintln(os.Stderr, "smith: "+msg) }
	// A variable Azure Pipelines has no value for (a fork's build, which gets
	// no secrets) arrives as its own macro text: it counts as unset.
	for _, name := range unexpandedMacros(os.Getenv, cacheEnvironment) {
		_ = os.Unsetenv(name)
	}
	secret := os.Getenv("SMITH_CACHE_KEY")
	if secret == "" {
		warn("the remote cache is off: SMITH_CACHE_KEY is not set, and nothing is shared unsigned")
		return local, nil
	}
	store, err := remotecache.Open(remote)
	if err != nil {
		warn("the remote cache is off: " + err.Error())
		return local, nil
	}
	readOnly := s.Config.Cache.RemoteReadOnly || os.Getenv("SMITH_CACHE_READ_ONLY") != "" || pullRequest(os.Getenv)
	return &remotecache.Tiered{Local: local, Remote: store, Secret: []byte(secret), ReadOnly: readOnly, Warn: warn}, nil
}

// Refresh re-reads files and rebuilds the graph, after a command changed the
// workspace (smith add).
func (s *Session) Refresh() error {
	files, err := s.Git.Files()
	if err != nil {
		return err
	}
	graph, err := discover(s.allowCycles, files, s.Plugins, s.Stacks, s.Config)
	if err != nil {
		return err
	}
	s.Files, s.Graph = files, graph
	return nil
}

// Reload re-reads smith.toml after a command wrote to it (smith add declares a
// recipe's stack there), then rebuilds the stack catalog and the graph from it.
// The plugins keep running: a change to [plugins] needs a new session.
func (s *Session) Reload() error {
	cfg, err := workspace.Load(s.Root)
	if err != nil {
		return err
	}
	catalog, err := pluginhost.NewCatalog(s.Plugins, cfg)
	if err != nil {
		return err
	}
	s.Config, s.Stacks = cfg, catalog
	return s.Refresh()
}

// Close stops the plugins.
func (s *Session) Close() {
	if s != nil && s.Plugins != nil {
		s.Plugins.Close()
	}
}
