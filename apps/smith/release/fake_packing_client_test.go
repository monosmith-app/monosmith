package release

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"sync"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// packingPlugin is a language plugin served in this process (no process is
// started). It writes the package files a test makes up and counts what it
// was asked.
type packingPlugin struct {
	name string
	root string
	// files gives the package files of a project and version: name to content.
	files func(project, version, salt string) map[string]string
	mu    sync.Mutex
	// salt changes the content of every file, to make a repack differ.
	salt string
	// packs counts the pack calls that wrote files; publishes counts the calls
	// that would have published to a registry.
	packs, publishes int
	// packDirs holds the PackDir of every pack call.
	packDirs []string
}

func (p *packingPlugin) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: p.name, Version: "1.0.0"}, Kind: protocol.KindLanguage, Manifests: []string{"x.json"}}, nil
}

func (p *packingPlugin) WriteVersion(protocol.VersionParams) (protocol.VersionResult, error) {
	return protocol.VersionResult{Files: []string{}}, nil
}

func (p *packingPlugin) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if params.PackDir == "" {
		p.publishes++
		return protocol.PublishResult{Published: true, Registry: "https://registry.npmjs.org/"}, nil
	}
	files := p.files(params.Project.Name, params.Version, p.salt)
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	var res protocol.PublishResult
	for _, n := range names {
		res.Assets = append(res.Assets, path.Join(params.PackDir, n))
	}
	p.packDirs = append(p.packDirs, params.PackDir)
	if params.DryRun {
		res.Skipped = "dry run"
		return res, nil
	}
	p.packs++
	dir := filepath.Join(p.root, filepath.FromSlash(params.PackDir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return res, err
	}
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(files[n]), 0o644); err != nil {
			return res, err
		}
	}
	return res, nil
}

// PacksPackages declares the packs capability.
func (p *packingPlugin) PacksPackages() {}

// counts reads the counters.
func (p *packingPlugin) counts() (packs, publishes int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.packs, p.publishes
}

// sha is the digest a checksum file states for a content.
func sha(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// packWorkspace is a git workspace with a bare origin, one package project
// per name, all with registry = "github-release", and the plugin that packs.
func packWorkspace(t *testing.T, plugin string, files func(project, version, salt string) map[string]string, projects ...string) (*session.Session, *packingPlugin) {
	t.Helper()
	g := repo(t)
	origin := gitclient.New(t.TempDir())
	if _, err := origin.Run("init", "--bare"); err != nil {
		t.Skip("git unavailable")
	}
	var list []*projectgraph.Project
	for _, name := range projects {
		commit(t, g, "packages/"+name+"/index.js", "feat: "+name)
		list = append(list, &projectgraph.Project{Project: protocol.Project{Name: name, Root: "packages/" + name, Kind: protocol.ProjectPackage, Version: "0.0.0", Exports: []string{name}, Registries: &protocol.Registries{PackToRelease: true}}, Plugin: plugin})
	}
	if _, err := g.Run("remote", "add", "origin", origin.Dir); err != nil {
		t.Fatal(err)
	}
	graph, err := projectgraph.New(list)
	if err != nil {
		t.Fatal(err)
	}
	fake := &packingPlugin{name: plugin, root: g.Dir, files: files}
	hostR, pluginW := io.Pipe()
	pluginR, hostW := io.Pipe()
	go func() { _ = serve.Serve(fake, pluginR, pluginW); _ = pluginW.Close() }()
	c, err := host.Connect(hostR, hostW, host.Options{WorkspaceRoot: g.Dir})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	cfg := workspace.Config{}
	cfg.Release.Remote, cfg.Release.GitHubReleases, cfg.Release.Branch = "origin", true, "main"
	return &session.Session{Root: g.Dir, Config: cfg, Git: g, Graph: graph, Plugins: &pluginhost.PluginSet{Clients: []*host.Client{c}}}, fake
}
